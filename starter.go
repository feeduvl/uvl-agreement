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
	router.HandleFunc("/hitec/agreement/annotationexport/", createAnnotationFromAgreement).Methods("POST")
	router.HandleFunc("/hitec/agreement/calculateKappa/", calculateKappaFromAgreement).Methods("POST")
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

// calculateKappaFromAgreement make and return the kappas
func calculateKappaFromAgreement(w http.ResponseWriter, r *http.Request) {
	var agreement Agreement
	err := json.NewDecoder(r.Body).Decode(&agreement)
	fmt.Printf("calculateKappaFromAgreement called: %s", agreement.Name)
	if err != nil {
		fmt.Printf("ERROR decoding body: %s, body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get all possible categories and relationship types for calculation of kappas
	toreCategories, err := RESTGetAllTores()
	handleErrorWithResponse(w, err, "ERROR retrieving all tore categories")
	toreRelationships, err := RESTGetAllRelationships()
	handleErrorWithResponse(w, err, "ERROR retrieving all relationships")

	// Get and parse kappas
	fleissKappa, brennanKappa := getKappas(agreement, toreCategories, toreRelationships)
	fmt.Printf("fleiss kappa is %v, brenan kappa is %v\n", fleissKappa, brennanKappa)
	var body = map[string]float64{}
	body["fleissKappa"] = fleissKappa
	body["brennanKappa"] = brennanKappa

	responseBody, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("Failed to marshal fleiss and brennan kappa")
	}
	w.Write(responseBody)
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

	// get all relevant annotation fields for an agreement
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
		fmt.Printf("\nAutomatically merge concurrent annotations\n")
		codeAlternatives = updateStatusOfCodeAlternatives(codeAlternatives, toreRelationships, len(annotationNames))
	}

	// parse the relevant fields into a struct
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

// createAnnotationFromAgreement create a new annotation from an agreement
func createAnnotationFromAgreement(w http.ResponseWriter, r *http.Request) {
	hasError := false
	var body map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&body)
	fmt.Printf("createAnnotationFromAgreement called: %s", createKeyValuePairs(body))
	if err != nil {
		fmt.Printf("ERROR decoding body: %s, body: %v\n", err, r.Body)
		hasError = true
		return
	}

	agreementName := body["agreementName"].(string)
	newAnnotationName := body["newAnnotationName"].(string)

	agreement, err := RESTGetAgreement(agreementName)
	handleErrorWithResponse(w, err, "ERROR retrieving annotation")

	// It should not be possible, if the agreement is not completed
	if agreement.IsCompleted == false {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Failure: Agreement is not completed."})
		return
	}
	newAnnotation := makeAnnotation(agreement, newAnnotationName)

	// parse result
	err = RESTPostAnnotation(newAnnotation)
	if err != nil {
		fmt.Printf("Failed to POST new annotation")
		hasError = true
		return
	}
	if hasError {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Annotation export failed!"})
	} else {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "New annotation created from agreement."})
	}
	return
}
