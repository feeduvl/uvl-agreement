package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	//"io"
	"log"
	"net/http"

	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/360EntSecGroup-Skylar/excelize"
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
	Docs   []DocWrapper `json:"docs" bson:"docs"`
	Tokens []Token      `json:"tokens" bson:"tokens"`

	TORECodeAlternatives     []TORECodeAlternatives     `json:"tore_code_alternatives" bson:"tore_code_alternatives"`
	WordCodeAlternatives     []WordCodeAlternatives     `json:"word_code_alternatives" bson:"word_code_alternatives"`
	RelationshipAlternatives []RelationshipAlternatives `json:"relationship_alternatives" bson:"relationship_alternatives"`
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

	docs, tokens, toreAlternatives, wordCodeAlternatives, relationshipAlternatives, err := initializeInfoFromAnnotations(w, annotationNames)
	if err != nil {
		fmt.Printf("Error getting annotations, returning")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	completeConcurrences := body["completeConcurrences"].(bool)
	fmt.Printf("CompleteConcurrences is set to %t", completeConcurrences)
	var completedToreAlternatives []TORECodeAlternatives

	if completeConcurrences {
		fmt.Printf("Automatically merge concurrent annotations")
		completedToreAlternatives = updateStatusOfToreCodeAlternatives(toreAlternatives, len(annotationNames))
	} else {
		completedToreAlternatives = toreAlternatives
	}

	fmt.Printf("completedToreAlternatives is set to %v", completedToreAlternatives)

	var relevantAgreementFields RelevantAgreementFields

	relevantAgreementFields.Docs = docs
	relevantAgreementFields.Tokens = tokens
	relevantAgreementFields.TORECodeAlternatives = completedToreAlternatives
	relevantAgreementFields.WordCodeAlternatives = wordCodeAlternatives
	relevantAgreementFields.RelationshipAlternatives = relationshipAlternatives

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
	[]TORECodeAlternatives,
	[]WordCodeAlternatives,
	[]RelationshipAlternatives,
	error,
) {
	var tores []TORECodeAlternatives
	var wordCodes []WordCodeAlternatives
	var relationships []RelationshipAlternatives
	var tokens []Token
	var docs []DocWrapper

	for i, annotationName := range annotationNames {
		annotation, err := RESTGetAnnotation(annotationName)
		handleErrorWithResponse(w, err, "ERROR retrieving annotation")
		if err != nil {
			return *new([]DocWrapper), *new([]Token), *new([]TORECodeAlternatives), *new([]WordCodeAlternatives), *new([]RelationshipAlternatives), err
		}

		log.Printf("Getting info from: " + annotationName)

		// Tokens and docs stay constant, so they can be filled with any annotation
		if i == 0 {
			tokens = annotation.Tokens
			docs = annotation.Docs
		}

		// Fill the alternatives individually with every single code
		for _, code := range annotation.Codes {

			if len(code.Tokens) != 0 {
				var toreCode = TORECodeAlternatives{
					AnnotationName: annotationName,
					MergeStatus:    "Pending",
					Tokens:         code.Tokens,
					Tore:           code.Tore,
				}
				tores = append(tores, toreCode)

				var wordCode = WordCodeAlternatives{
					AnnotationName: annotationName,
					MergeStatus:    "Pending",
					Tokens:         code.Tokens,
					Name:           code.Name,
				}
				wordCodes = append(wordCodes, wordCode)

				var relationshipCode = RelationshipAlternatives{
					AnnotationName:          annotationName,
					MergeStatus:             "Pending",
					Tokens:                  code.Tokens,
					RelationshipMemberships: code.RelationshipMemberships,
					TORERelationships:       annotation.TORERelationships,
				}
				relationships = append(relationships, relationshipCode)
			}
		}

	}

	return docs, tokens, tores, wordCodes, relationships, nil
}

type MergeCandidate struct {
	Tokens                    []*int
	Tore                      string
	annotationNameOccurrences []string
}

func updateStatusOfToreCodeAlternatives(
	toreAlternatives []TORECodeAlternatives,
	numberOfAnnotations int,
) []TORECodeAlternatives {
	var mergeCandidates []MergeCandidate
	var rejected [][]*int
	for _, tore := range toreAlternatives {
		if len(mergeCandidates) == 0 {
			if len(rejected) == 0 {
				mergeCandidates = append(mergeCandidates, MergeCandidate{tore.Tokens, tore.Tore, []string{tore.AnnotationName}})
			} else {
				mergeCandidates, rejected = testRejection(tore, mergeCandidates, rejected)
			}
		} else {
			for i, candidate := range mergeCandidates {
				if testEqSlice(tore.Tokens, candidate.Tokens) {
					if tore.Tore != candidate.Tore {
						rejected = append(rejected, candidate.Tokens)
						mergeCandidates = append(mergeCandidates[:i], mergeCandidates[i+1:]...)
					} else {
						var isNew = true
						for _, annoNameOccurrence := range candidate.annotationNameOccurrences {
							if annoNameOccurrence == tore.AnnotationName {
								isNew = false
							}
						}
						if isNew {
							mergeCandidates[i].annotationNameOccurrences = append(mergeCandidates[i].annotationNameOccurrences, tore.AnnotationName)
						}
					}
				} else {
					mergeCandidates, rejected = testRejection(tore, mergeCandidates, rejected)
				}
			}
		}

	}
	return setMergeStatus(toreAlternatives, mergeCandidates, numberOfAnnotations)
}

func setMergeStatus(
	toreAlternatives []TORECodeAlternatives,
	mergeCandidates []MergeCandidate,
	numberOfAnnotations int,
) []TORECodeAlternatives {

	for _, candidate := range mergeCandidates {
		if len(candidate.annotationNameOccurrences) == numberOfAnnotations {
			var isAccepted = false
			for i, tore := range toreAlternatives {
				if !isAccepted {
					if testEqSlice(candidate.Tokens, tore.Tokens) {
						toreAlternatives[i].MergeStatus = "Accepted"
						isAccepted = true
					}
				} else {
					if testEqSlice(candidate.Tokens, tore.Tokens) {
						toreAlternatives[i].MergeStatus = "Declined"
					}
				}
			}
		}
	}
	return toreAlternatives
}

func testRejection(
	tore TORECodeAlternatives,
	mergeCandidates []MergeCandidate,
	rejected [][]*int,
) ([]MergeCandidate, [][]*int) {

	var isAReject = false
	for _, reject := range rejected {
		if testEqSlice(tore.Tokens, reject) {
			isAReject = true
			rejected = append(rejected, tore.Tokens)
			for i, candidate := range mergeCandidates {
				if testEqSlice(tore.Tokens, candidate.Tokens) {
					mergeCandidates = append(mergeCandidates[:i], mergeCandidates[i+1:]...)
				}
			}
		}
	}
	if !isAReject {
		mergeCandidates = append(mergeCandidates, MergeCandidate{tore.Tokens, tore.Tore, []string{tore.AnnotationName}})
	}
	return mergeCandidates, rejected
}

func testEqSlice(a, b []*int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if *a[i] != *b[i] {
			return false
		}
	}
	return true
}
