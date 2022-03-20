package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	//"io"
	"log"
	"net/http"

	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	log.SetOutput(os.Stdout)
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	router := makeRouter()

	fmt.Println("uvl-agreement MS running")
	log.Fatal(http.ListenAndServe(":9662", handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router)))
}

func makeRouter() *mux.Router {
	router := mux.NewRouter()

	// Init
	router.HandleFunc("/hitec/agreement/annotationinfo/", getInfoFromAnnotations).Methods("POST")
	return router
}

func handleErrorWithResponse(w http.ResponseWriter, err error, message string) {
	if err != nil {
		_ = json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: message})
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}
}

func createKeyValuePairs(m map[string]interface{}) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%#v\"\n", key, value)
	}
	return b.String()
}

type RelevantAgreementFields struct {
	Docs              []DocWrapper       `json:"docs" bson:"docs"`
	Tokens            []Token            `json:"tokens" bson:"tokens"`
	TORERelationships []TORERelationship `json:"tore_relationships" bson:"tore_relationships"`

	CodeAlternatives []CodeAlternatives `json:"code_alternatives" bson:"code_alternatives"`
}

// getInfoFromAnnotations make and return the alternatives, tokens and docs for agreement
func getInfoFromAnnotations(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&body)
	fmt.Printf("getInfoFromAnnotations called: %s", createKeyValuePairs(body))
	if err != nil {
		fmt.Printf("ERROR decoding body: %s, body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var annotationNames []string
	bodyAnnotationNames := body["annotationNames"].([]interface{})
	for _, value := range bodyAnnotationNames {
		fmt.Printf("element: %v\n", value)
		annotationNames = append(annotationNames, value.(string))
	}

	docs, tokens, toreRelationships, codeAlternatives, err := initializeInfoFromAnnotations(w, annotationNames)
	if err != nil {
		fmt.Printf("Error getting annotations, returning")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	completeConcurrences := body["completeConcurrences"].(bool)
	fmt.Printf("CompleteConcurrences is set to %t", completeConcurrences)

	if completeConcurrences {
		fmt.Printf("Automatically merge concurrent annotations")
		codeAlternatives = updateStatusOfCodeAlternatives(codeAlternatives, toreRelationships, len(annotationNames))
	}

	var relevantAgreementFields RelevantAgreementFields

	relevantAgreementFields.Docs = docs
	relevantAgreementFields.Tokens = tokens
	relevantAgreementFields.TORERelationships = toreRelationships
	relevantAgreementFields.CodeAlternatives = codeAlternatives

	finalRelevantFields, err := json.Marshal(relevantAgreementFields)
	if err != nil {
		fmt.Printf("Failed to marshal relevantAgreementFields")
	}
	w.Write(finalRelevantFields)
}

// postAgreementTokenize Tokenize a document and return the result
func initializeInfoFromAnnotations(
	w http.ResponseWriter, annotationNames []string,
) (
	[]DocWrapper,
	[]Token,
	[]TORERelationship,
	[]CodeAlternatives,
	error,
) {
	var codes []CodeAlternatives
	var tokens []Token
	var docs []DocWrapper
	var toreRelationships []TORERelationship

	var indexCounter = 0
	var relationshipIndexCounter = 0

	for i, annotationName := range annotationNames {
		annotation, err := RESTGetAnnotation(annotationName)
		handleErrorWithResponse(w, err, "ERROR retrieving annotation")
		if err != nil {
			return *new([]DocWrapper), *new([]Token), *new([]TORERelationship), *new([]CodeAlternatives), err
		}

		log.Printf("Getting info from: " + annotationName)

		// Relevant fields of Tokens and docs stay constant, so they can be filled with any annotation
		if i == 0 {
			tokens = annotation.Tokens
			docs = annotation.Docs
		}

		// Necessary to get global ToreRelationships
		var numberOfRelationshipsInAnnotation = len(annotation.TORERelationships)
		for i, _ := range annotation.TORERelationships {
			*annotation.TORERelationships[i].Index += relationshipIndexCounter
		}

		toreRelationships = append(toreRelationships, annotation.TORERelationships...)

		// Fill the alternatives individually with every single code
		for _, code := range annotation.Codes {
			fmt.Printf("Next is assigning index to indexcounter\n")
			*code.Index = indexCounter
			for i, _ := range code.RelationshipMemberships {
				fmt.Printf("Next is assigning index to relationshipmemberships\n")
				*code.RelationshipMemberships[i] += relationshipIndexCounter
				for j, toreRel := range toreRelationships {
					if toreRel.TOREEntity != nil {
						fmt.Printf("Next istesting index of membership to toreRelIndex: %v\n", *toreRel.Index)
						if *code.RelationshipMemberships[i] == *toreRel.Index {
							fmt.Printf("Next istesting reassigning toreEntity: %v\n", *toreRelationships[j].TOREEntity)
							*toreRelationships[j].TOREEntity = indexCounter
						}
					}
				}
			}
			if len(code.Tokens) != 0 {
				var code = CodeAlternatives{
					Index:          indexCounter,
					AnnotationName: annotationName,
					MergeStatus:    "Pending",
					Code:           code,
				}
				codes = append(codes, code)
				indexCounter++
			}
		}
		relationshipIndexCounter += numberOfRelationshipsInAnnotation

	}

	return docs, tokens, toreRelationships, codes, nil
}
