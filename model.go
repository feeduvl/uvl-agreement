package main

import (
	"time"
)

type DocWrapper struct {
	Name       string `json:"name" bson:"name"`
	BeginIndex *int   `json:"begin_index" bson:"begin_index"`
	EndIndex   *int   `json:"end_index" bson:"end_index"`
}

type TORERelationship struct {
	TOREEntity       *int   `json:"TOREEntity" bson:"TOREEntity"`
	TargetTokens     []*int `json:"target_tokens" bson:"target_tokens"`
	RelationshipName string `json:"relationship_name" bson:"relationship_name"`
	Index            *int   `json:"index" bson:"index"`
}

type Code struct {
	Tokens                  []*int `json:"tokens" bson:"tokens"`
	Name                    string `json:"name" bson:"name"`
	Tore                    string `json:"tore" bson:"tore"`
	Index                   *int   `json:"index" bson:"index"`
	RelationshipMemberships []*int `json:"relationship_memberships" bson:"relationship_memberships"`
}

type Token struct {
	Index        *int   `json:"index" bson:"index"`
	Name         string `validate:"nonzero" json:"name" bson:"name"`
	Lemma        string `validate:"nonzero" json:"lemma" bson:"lemma"`
	Pos          string `validate:"nonzero" json:"pos" bson:"pos"`
	NumNameCodes int    `json:"num_name_codes" bson:"num_name_codes"`
	NumToreCodes int    `json:"num_tore_codes" bson:"num_tore_codes"`
}

type Annotation struct {
	UploadedAt  time.Time `validate:"nonzero" json:"uploaded_at" bson:"uploaded_at"`
	LastUpdated time.Time `json:"last_updated" bson:"last_updated"`

	Name    string `validate:"nonzero" json:"name" bson:"name"`
	Dataset string `validate:"nonzero" json:"dataset" bson:"dataset"`

	Docs              []DocWrapper       `json:"docs" bson:"docs"`
	Tokens            []Token            `json:"tokens" bson:"tokens"`
	Codes             []Code             `json:"codes" bson:"codes"`
	TORERelationships []TORERelationship `json:"tore_relationships" bson:"tore_relationships"`
}

// AgreementStatistics model, the initial and current kappas. Name is unique
type AgreementStatistics struct {
	KappaName    string  `validate:"nonzero" json:"kappa_name" bson:"kappa_name"`
	InitialKappa float64 `json:"initial_kappa" bson:"initial_kappa"`
	CurrentKappa float64 `json:"current_kappa" bson:"current_kappa"`
}

// CodeAlternatives model, shows all code alternatives from all annotations, MergeStatus can be set to Pending, Accepted or Declined
type CodeAlternatives struct {
	AnnotationName string `json:"annotation_name" bson:"annotation_name"`
	MergeStatus    string `validate:"nonzero" json:"merge_status" bson:"merge_status"`
	Index          int    `json:"index" bson:"index"`

	Code Code `json:"code" bson:"code"`
}

// Agreement model
type Agreement struct {
	CreatedAt   time.Time `validate:"nonzero" json:"created_at" bson:"created_at"`
	LastUpdated time.Time `json:"last_updated" bson:"last_updated"`

	Name        string   `validate:"nonzero" json:"name" bson:"name"`
	Dataset     string   `validate:"nonzero" json:"dataset" bson:"dataset"`
	Annotations []string `json:"annotation_names" bson:"annotation_names"`

	Docs              []DocWrapper       `json:"docs" bson:"docs"`
	Tokens            []Token            `json:"tokens" bson:"tokens"`
	TORERelationships []TORERelationship `json:"tore_relationships" bson:"tore_relationships"`

	CodeAlternatives    []CodeAlternatives    `json:"code_alternatives" bson:"code_alternatives"`
	AgreementStatistics []AgreementStatistics `json:"agreement_statistics" bson:"agreement_statistics"`

	IsCompleted bool `json:"is_completed" bson:"is_completed"`
}

// ResponseMessage model
type ResponseMessage struct {
	Message string `json:"message"`
	Status  bool   `json:"status"`
}
