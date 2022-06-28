package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"log"
)

var baseURL = os.Getenv("BASE_URL")
var bearerToken = "Bearer " + os.Getenv("BEARER_TOKEN")

const (
	// agreement
	endpointGetAnnotation       = "/hitec/repository/concepts/annotation/name/"
	endpointGetAgreement        = "/hitec/repository/concepts/agreement/name/"
	endpointPostStoreAnnotation = "/hitec/repository/concepts/store/annotation/"
	endpointGetAllTores         = "/hitec/repository/concepts/annotation/tores"
	endpointGetAllRelationships = "/hitec/repository/concepts/annotation/relationships"

	GET           = "GET"
	POST          = "POST"
	AUTHORIZATION = "Authorization"
	ACCEPT        = "Accept"
	TYPE_JSON     = "application/json"

	errJsonMessageTemplate = "ERR - json formatting error: %v\n"
)

var client = getHTTPClient()

func getHTTPClient() *http.Client {
	pwd, _ := os.Getwd()
	caCert, err := ioutil.ReadFile(pwd + "/ca_chain.crt")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	timeout := 15 * time.Minute

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
				// InsecureSkipVerify: true,
			},
		},
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			req.Header.Add(AUTHORIZATION, bearerToken)
			return nil
		},
	}

	return client
}

func createRequest(method string, url string, payload io.Reader) (*http.Request, error) {
	req, _ := http.NewRequest(method, url, payload)
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	return req, nil
}

// RESTGetAnnotation returns annotation, err
func RESTGetAnnotation(annotationName string) (Annotation, error) {
	requestBody := new(bytes.Buffer)
	var annotation Annotation

	// make request
	url := baseURL + endpointGetAnnotation + url.QueryEscape(annotationName)
	req, _ := createRequest(GET, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR get annotation %v\n", err)
		return annotation, err
	}
	// parse result
	err = json.NewDecoder(res.Body).Decode(&annotation)
	if err != nil {
		log.Printf("ERR parsing annotation %v\n", err)
		return annotation, err
	}
	return annotation, err
}

// RESTGetAgreement returns agreement, err
func RESTGetAgreement(agreementName string) (Agreement, error) {
	requestBody := new(bytes.Buffer)
	var agreement Agreement

	// make request
	url := baseURL + endpointGetAgreement + url.QueryEscape(agreementName)
	req, _ := createRequest(GET, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR get agreement %v\n", err)
		return agreement, err
	}
	// parse result
	err = json.NewDecoder(res.Body).Decode(&agreement)
	if err != nil {
		log.Printf("ERR parsing agreement %v\n", err)
		log.Printf("this is the result: %v\n", res.Body)
		return agreement, err
	}
	return agreement, err
}

// RESTPostStoreAnnotation returns err
func RESTPostStoreAnnotation(annotation Annotation) error {
	requestBody := new(bytes.Buffer)
	_ = json.NewEncoder(requestBody).Encode(annotation)
	url := baseURL + endpointPostStoreAnnotation
	req, _ := createRequest(POST, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR post store annotation %v\n", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	return nil
}

// RESTGetAllTores returns toreCategories, err
func RESTGetAllTores() (ToreCategories, error) {
	requestBody := new(bytes.Buffer)
	var toreCategories ToreCategories

	// make request
	url := baseURL + endpointGetAllTores
	req, _ := createRequest(GET, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR get tores %v\n", err)
		return toreCategories, err
	}
	// parse result
	err = json.NewDecoder(res.Body).Decode(&toreCategories)
	if err != nil {
		log.Printf("ERR parsing toreCategories %v\n", err)
		return toreCategories, err
	}
	return toreCategories, err
}

// RESTGetAllRelationships returns toreRelationships, err
func RESTGetAllRelationships() (ToreRelationships, error) {
	requestBody := new(bytes.Buffer)
	var toreRelationships ToreRelationships

	// make request
	url := baseURL + endpointGetAllRelationships
	req, _ := createRequest(GET, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR get relationships %v\n", err)
		return toreRelationships, err
	}
	// parse result
	err = json.NewDecoder(res.Body).Decode(&toreRelationships)
	if err != nil {
		log.Printf("ERR parsing toreRelationships %v\n", err)
		return toreRelationships, err
	}
	return toreRelationships, err
}
