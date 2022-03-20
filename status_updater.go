package main

import (
	"fmt"
)

type CodeMergeCandidate struct {
	Tokens                    []*int
	Name                      string
	Tore                      string
	RelationshipMemberships   []*int
	annotationNameOccurrences []string
}

func updateStatusOfCodeAlternatives(
	codeAlternatives []CodeAlternatives,
	toreRelationships []TORERelationship,
	numberOfAnnotations int,
) []CodeAlternatives {
	var mergeCandidates []CodeMergeCandidate
	var rejected [][]*int
	for _, codeAlternative := range codeAlternatives {
		if len(mergeCandidates) == 0 {
			if len(rejected) == 0 {
				// First candidate is automatically added to mergeCandidates
				var newCandidate = CodeMergeCandidate{
					codeAlternative.Code.Tokens,
					codeAlternative.Code.Name,
					codeAlternative.Code.Tore,
					codeAlternative.Code.RelationshipMemberships,
					[]string{codeAlternative.AnnotationName},
				}
				mergeCandidates = append(mergeCandidates, newCandidate)
			} else {
				// Test, if candidate is already rejected
				// If yes, nothing happens. If no, add candidate to mergeCandidates
				mergeCandidates = testCodeRejection(codeAlternative, mergeCandidates, rejected)
			}
		} else {
			var isFound = false
			for i, candidate := range mergeCandidates {
				// Candidate is already in mergeCandidates
				if testEqSlice(codeAlternative.Code.Tokens, candidate.Tokens) {
					isFound = true
					// When any property is changed, it is added to rejected, and removed from mergeCandidates
					if (codeAlternative.Code.Tore != candidate.Tore) || (codeAlternative.Code.Name != candidate.Name) || (!testRelationshipsAreEqual(codeAlternative.Code.RelationshipMemberships, candidate.RelationshipMemberships, toreRelationships)) {
						rejected = append(rejected, candidate.Tokens)
						mergeCandidates = append(mergeCandidates[:i], mergeCandidates[i+1:]...)
						break
					} else {
						// if nothing has changed, the annotationName is added
						var isNew = true
						for _, annoNameOccurrence := range candidate.annotationNameOccurrences {
							if annoNameOccurrence == codeAlternative.AnnotationName {
								isNew = false
							}
						}
						if isNew {
							mergeCandidates[i].annotationNameOccurrences = append(mergeCandidates[i].annotationNameOccurrences, codeAlternative.AnnotationName)
							break
						}
					}
				}
			}
			// Candidate is not found in mergeCandidates, so either it is new, or it is already rejected
			if !isFound {
				mergeCandidates = testCodeRejection(codeAlternative, mergeCandidates, rejected)
			}
		}

	}
	return setCodeMergeStatus(codeAlternatives, mergeCandidates, numberOfAnnotations)
}

func setCodeMergeStatus(
	codeAlternatives []CodeAlternatives,
	mergeCandidates []CodeMergeCandidate,
	numberOfAnnotations int,
) []CodeAlternatives {

	for _, candidate := range mergeCandidates {
		if len(candidate.annotationNameOccurrences) == numberOfAnnotations {
			var isAccepted = false
			for i, codeAlternative := range codeAlternatives {
				if !isAccepted {
					if testEqSlice(candidate.Tokens, codeAlternative.Code.Tokens) {
						codeAlternatives[i].MergeStatus = "Accepted"
						isAccepted = true
					}
				} else {
					if testEqSlice(candidate.Tokens, codeAlternative.Code.Tokens) {
						codeAlternatives[i].MergeStatus = "Declined"
					}
				}
			}
		}
	}
	return codeAlternatives
}

func testCodeRejection(
	codeAlternative CodeAlternatives,
	mergeCandidates []CodeMergeCandidate,
	rejected [][]*int,
) []CodeMergeCandidate {

	var isAReject = false
	for _, reject := range rejected {
		// if candidate is found in rejected
		if testEqSlice(codeAlternative.Code.Tokens, reject) {
			isAReject = true
			for i, candidate := range mergeCandidates {
				if testEqSlice(codeAlternative.Code.Tokens, candidate.Tokens) {
					mergeCandidates = append(mergeCandidates[:i], mergeCandidates[i+1:]...)
					break
				}
			}
			break
		}
	}
	if !isAReject {
		fmt.Printf("The candidate is not found in rejected\n")
		var newCandidate = CodeMergeCandidate{
			codeAlternative.Code.Tokens,
			codeAlternative.Code.Name,
			codeAlternative.Code.Tore,
			codeAlternative.Code.RelationshipMemberships,
			[]string{codeAlternative.AnnotationName},
		}
		mergeCandidates = append(mergeCandidates, newCandidate)
	}
	return mergeCandidates
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

func testRelationshipsAreEqual(
	a []*int,
	b []*int,
	relationships []TORERelationship,
) bool {
	if len(a) != len(b) {
		return false
	}
	fmt.Printf("Reached testRelationshipsAreEqual, a is %v, b is %v \n", a, b)
	var relationshipsA []TORERelationship
	var relationshipsB []TORERelationship
	for _, relationshipIndex := range a {
		for _, relationship := range relationships {
			if relationshipIndex == relationship.Index {
				relationshipsA = append(relationshipsA, relationship)
			}
		}
	}
	for _, relationshipIndex := range b {
		for _, relationship := range relationships {
			if relationshipIndex == relationship.Index {
				relationshipsB = append(relationshipsB, relationship)
			}
		}
	}
	fmt.Printf("individual relationships loaded a is %v, b is %v \n", relationshipsA, relationshipsB)
	for _, rel1 := range relationshipsA {
		var indicesToRemove []int
		for j, rel2 := range relationshipsB {
			if rel1.RelationshipName == rel2.RelationshipName && testEqSlice(rel1.TargetTokens, rel2.TargetTokens) {
				indicesToRemove = append(indicesToRemove, j)
			}
		}
		for _, index := range indicesToRemove {
			relationshipsB = append(relationshipsB[:index], relationshipsB[index+1:]...)
		}
		indicesToRemove = []int{}
	}
	return len(relationshipsB) == 0
}
