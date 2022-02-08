package main

import "fmt"

type CodeMergeCandidate struct {
	Tokens                    []*int
	Name                      string
	Tore                      string
	RelationshipMemberships   []*int
	annotationNameOccurrences []string
}

func updateStatusOfCodeAlternatives(
	codeAlternatives []CodeAlternatives,
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
			for i, candidate := range mergeCandidates {
				fmt.Printf(" \n len(mergecandidate): %v \n", len(mergeCandidates))
				fmt.Printf("mergecandidate: %v \n", mergeCandidates)
				fmt.Printf("codeAlternative: %v \n", codeAlternative)
				fmt.Printf("rejected: %v \n", rejected)
				var isFound = false
				// Candidate is already in mergeCandidates
				if testEqSlice(codeAlternative.Code.Tokens, candidate.Tokens) {
					fmt.Printf("The candidate is found \n")
					isFound = true
					// When any property is changed, it is added to rejected, and removed from mergeCandidates
					if (codeAlternative.Code.Tore != candidate.Tore) || (codeAlternative.Code.Name != candidate.Name) || (!testEqSlice(codeAlternative.Code.RelationshipMemberships, candidate.RelationshipMemberships)) {
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
				// Candidate is not found in mergeCandidates, so either it is new, or it is already rejected
				if !isFound {
					fmt.Printf("The candidate is not found \n")
					mergeCandidates = testCodeRejection(codeAlternative, mergeCandidates, rejected)
				}
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
			fmt.Printf("The candidate is found in rejected \n")
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
