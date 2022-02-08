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
	fmt.Printf("Number of codeAlternatives: %v\n\n", len(codeAlternatives))
	for _, codeAlternative := range codeAlternatives {
		fmt.Printf("Searching through codeAlternative %v\n\n", codeAlternative)
		if len(mergeCandidates) == 0 {
			if len(rejected) == 0 {
				var newCandidate = CodeMergeCandidate{
					codeAlternative.Code.Tokens,
					codeAlternative.Code.Name,
					codeAlternative.Code.Tore,
					codeAlternative.Code.RelationshipMemberships,
					[]string{codeAlternative.AnnotationName},
				}
				mergeCandidates = append(mergeCandidates, newCandidate)
			} else {
				mergeCandidates, rejected = testCodeRejection(codeAlternative, mergeCandidates, rejected)
			}
		} else {
			for i, candidate := range mergeCandidates {
				//fmt.Printf("Number of candidates %v\n\n", len(mergeCandidates))
				//fmt.Printf("Comparing to candidate %v\n\n", candidate)
				if testEqSlice(codeAlternative.Code.Tokens, candidate.Tokens) {
					if (codeAlternative.Code.Tore != candidate.Tore) || (codeAlternative.Code.Name != candidate.Name) || (!testEqSlice(codeAlternative.Code.RelationshipMemberships, candidate.RelationshipMemberships)) {
						//fmt.Printf("%v already found\n\n", codeAlternative)
						rejected = append(rejected, candidate.Tokens)
						mergeCandidates = append(mergeCandidates[:i], mergeCandidates[i+1:]...)
					} else {
						var isNew = true
						for _, annoNameOccurrence := range candidate.annotationNameOccurrences {
							if annoNameOccurrence == codeAlternative.AnnotationName {
								isNew = false
							}
						}
						if isNew {
							//fmt.Printf("%v new, is added\n\n", codeAlternative)
							mergeCandidates[i].annotationNameOccurrences = append(mergeCandidates[i].annotationNameOccurrences, codeAlternative.AnnotationName)
							break
						}
					}
				} else {
					mergeCandidates, rejected = testCodeRejection(codeAlternative, mergeCandidates, rejected)
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
		fmt.Printf("Setting status of MergeCandidate %v\n\n", candidate)
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
) ([]CodeMergeCandidate, [][]*int) {

	var isAReject = false
	for _, reject := range rejected {
		if testEqSlice(codeAlternative.Code.Tokens, reject) {
			isAReject = true
			rejected = append(rejected, codeAlternative.Code.Tokens)
			for i, candidate := range mergeCandidates {
				if testEqSlice(codeAlternative.Code.Tokens, candidate.Tokens) {
					mergeCandidates = append(mergeCandidates[:i], mergeCandidates[i+1:]...)
				}
			}
		}
	}
	if !isAReject {
		var newCandidate = CodeMergeCandidate{
			codeAlternative.Code.Tokens,
			codeAlternative.Code.Name,
			codeAlternative.Code.Tore,
			codeAlternative.Code.RelationshipMemberships,
			[]string{codeAlternative.AnnotationName},
		}
		mergeCandidates = append(mergeCandidates, newCandidate)
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
