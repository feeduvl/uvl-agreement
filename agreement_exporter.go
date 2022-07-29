package main

import (
	"time"
)

func makeAnnotation(agreement Agreement, newAnnotationName string) Annotation {
	toreRelationships := agreement.TORERelationships
	codeAlternatives := agreement.CodeAlternatives

	acceptedCodes, acceptedToreRelationships := makeAcceptedToreRelationshipsAndCodes(toreRelationships, codeAlternatives)
	updatedTokens := updateTokens(agreement, acceptedCodes)

	var newAnnotation = Annotation{
		UploadedAt:        time.Now(),
		LastUpdated:       time.Now(),
		Name:              newAnnotationName,
		Dataset:           agreement.Dataset,
		Docs:              agreement.Docs,
		Tokens:            updatedTokens,
		Codes:             acceptedCodes,
		TORERelationships: acceptedToreRelationships,
	}

	return newAnnotation
}

// The fields numNameCodes and numToreCodes are not used in the agreement, for are necessary for annotations
// So they have to be adapted before the export
func updateTokens(agreement Agreement, acceptedCodes []Code) []Token {
	var newTokens []Token
	for _, token := range agreement.Tokens {
		numNameCodes := 0
		numToreCodes := 0
		// Change number of names and tores
		for _, acceptedCode := range acceptedCodes {
			for _, tokenInCode := range acceptedCode.Tokens {
				if *tokenInCode == *token.Index {
					if acceptedCode.Name != "" {
						numNameCodes++
					}
					if acceptedCode.Tore != "" {
						numToreCodes++
					}
				}
			}
		}
		var newToken = Token{
			Index:        token.Index,
			Name:         token.Name,
			Lemma:        token.Lemma,
			Pos:          token.Pos,
			NumNameCodes: numNameCodes,
			NumToreCodes: numToreCodes,
		}
		newTokens = append(newTokens, newToken)
	}
	return newTokens
}

// Only accepted codes are used in the annotation, so all codes and relationships that are not accepted have to be removed
// The index of codes and relationships has to be adapted as well
func makeAcceptedToreRelationshipsAndCodes(
	toreRelationships []TORERelationship,
	codeAlternatives []CodeAlternatives,
) ([]Code, []TORERelationship) {
	codeIndex := 0
	var acceptedCodes []Code
	var acceptedToreRelationships []TORERelationship

	for i, codeAlternative := range codeAlternatives {
		if codeAlternative.MergeStatus == "Accepted" {
			*codeAlternatives[i].Code.Index = codeIndex
			for _, usedRelIndex := range codeAlternative.Code.RelationshipMemberships {
				for j, toreRel := range toreRelationships {
					if toreRel.TOREEntity != nil {
						if *usedRelIndex == *toreRel.Index {
							*toreRelationships[j].TOREEntity = codeIndex
							acceptedToreRelationships = append(acceptedToreRelationships, toreRel)
							break
						}
					}
				}
			}
			codeAlternatives[i].Code.RelationshipMemberships = []*int{}
			acceptedCodes = append(acceptedCodes, codeAlternatives[i].Code)
			codeIndex++
		}
	}

	toreRelIndex := 0
	for i, acceptedRel := range acceptedToreRelationships {
		*acceptedToreRelationships[i].Index = toreRelIndex
		for j, acceptedCode := range acceptedCodes {
			if *acceptedCode.Index == *acceptedRel.TOREEntity {
				acceptedCodes[j].RelationshipMemberships = append(acceptedCodes[j].RelationshipMemberships, acceptedToreRelationships[i].Index)
				break
			} else {
			}
		}
		toreRelIndex++
	}
	return acceptedCodes, acceptedToreRelationships
}
