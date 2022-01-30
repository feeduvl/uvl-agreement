package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"log"
)

var baseURL = os.Getenv("BASE_URL")
var bearerToken = "Bearer " + os.Getenv("BEARER_TOKEN")

const (
	// agreement
	endpointGetAnnotation = "/hitec/repository/concepts/annotation/name/"

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
	url := baseURL + endpointGetAnnotation + annotationName
	req, _ := createRequest(GET, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR get annotation %v\n", err)
		return annotation, err
	}
	// parse result
	err = json.NewDecoder(res.Body).Decode(&annotation)
	if err != nil {
		log.Printf("ERR parsing dataset %v\n", err)
		return annotation, err
	}
	return annotation, err
}