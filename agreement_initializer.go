package main

import (
	"log"
	"net/http"
)

// RelevantAgreementFields Used to group relevant agreement fields
type RelevantAgreementFields struct {
	Docs              []DocWrapper       `json:"docs" bson:"docs"`
	Tokens            []Token            `json:"tokens" bson:"tokens"`
	TORERelationships []TORERelationship `json:"tore_relationships" bson:"tore_relationships"`

	CodeAlternatives []CodeAlternatives `json:"code_alternatives" bson:"code_alternatives"`
}

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
		for i, toreRel := range annotation.TORERelationships {
			if toreRel.TOREEntity != nil && toreRel.Index != nil {
				*annotation.TORERelationships[i].Index += relationshipIndexCounter
			}
		}

		toreRelationships = append(toreRelationships, annotation.TORERelationships...)

		// Fill the alternatives individually with every single code, set all codes to pending
		for _, code := range annotation.Codes {
			for i, _ := range code.RelationshipMemberships {
				*code.RelationshipMemberships[i] += relationshipIndexCounter
				for j, toreRel := range toreRelationships {
					if toreRel.TOREEntity != nil && toreRel.Index != nil {
						if *code.RelationshipMemberships[i] == *toreRel.Index {
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
