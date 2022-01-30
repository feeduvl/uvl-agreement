package main

type ToreCodeMergeCandidate struct {
	Tokens                    []*int
	Tore                      string
	annotationNameOccurrences []string
}

type WordCodeMergeCandidate struct {
	Tokens                    []*int
	Name                      string
	annotationNameOccurrences []string
}

type RelationshipMergeCandidate struct {
	Tokens                    []*int
	RelationshipMemberships   []*int
	annotationNameOccurrences []string
}

func updateStatusOfToreCodeAlternatives(
	toreAlternatives []TORECodeAlternatives,
	numberOfAnnotations int,
) []TORECodeAlternatives {
	var mergeCandidates []ToreCodeMergeCandidate
	var rejected [][]*int
	for _, tore := range toreAlternatives {
		if len(mergeCandidates) == 0 {
			if len(rejected) == 0 {
				mergeCandidates = append(mergeCandidates, ToreCodeMergeCandidate{tore.Tokens, tore.Tore, []string{tore.AnnotationName}})
			} else {
				mergeCandidates, rejected = testToreCodeRejection(tore, mergeCandidates, rejected)
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
					mergeCandidates, rejected = testToreCodeRejection(tore, mergeCandidates, rejected)
				}
			}
		}

	}
	return setToreCodeMergeStatus(toreAlternatives, mergeCandidates, numberOfAnnotations)
}

func updateStatusOfWordCodeAlternatives(
	wordCodeAlternatives []WordCodeAlternatives,
	numberOfAnnotations int,
) []WordCodeAlternatives {
	var mergeCandidates []WordCodeMergeCandidate
	var rejected [][]*int
	for _, wordCode := range wordCodeAlternatives {
		if len(mergeCandidates) == 0 {
			if len(rejected) == 0 {
				mergeCandidates = append(mergeCandidates, WordCodeMergeCandidate{wordCode.Tokens, wordCode.Name, []string{wordCode.AnnotationName}})
			} else {
				mergeCandidates, rejected = testWordCodeRejection(wordCode, mergeCandidates, rejected)
			}
		} else {
			for i, candidate := range mergeCandidates {
				if testEqSlice(wordCode.Tokens, candidate.Tokens) {
					if wordCode.Name != candidate.Name {
						rejected = append(rejected, candidate.Tokens)
						mergeCandidates = append(mergeCandidates[:i], mergeCandidates[i+1:]...)
					} else {
						var isNew = true
						for _, annoNameOccurrence := range candidate.annotationNameOccurrences {
							if annoNameOccurrence == wordCode.AnnotationName {
								isNew = false
							}
						}
						if isNew {
							mergeCandidates[i].annotationNameOccurrences = append(mergeCandidates[i].annotationNameOccurrences, wordCode.AnnotationName)
						}
					}
				} else {
					mergeCandidates, rejected = testWordCodeRejection(wordCode, mergeCandidates, rejected)
				}
			}
		}

	}
	return setWordCodeMergeStatus(wordCodeAlternatives, mergeCandidates, numberOfAnnotations)
}

func updateStatusOfRelationshipAlternatives(
	relationshipAlternatives []RelationshipAlternatives,
	numberOfAnnotations int,
) []RelationshipAlternatives {
	var mergeCandidates []RelationshipMergeCandidate
	var rejected [][]*int
	for _, relationship := range relationshipAlternatives {
		if len(mergeCandidates) == 0 {
			if len(rejected) == 0 {
				mergeCandidates = append(mergeCandidates, RelationshipMergeCandidate{relationship.Tokens, relationship.RelationshipMemberships, []string{relationship.AnnotationName}})
			} else {
				mergeCandidates, rejected = testRelationshipRejection(relationship, mergeCandidates, rejected)
			}
		} else {
			for i, candidate := range mergeCandidates {
				if testEqSlice(relationship.Tokens, candidate.Tokens) {
					if !testEqSlice(relationship.RelationshipMemberships, candidate.RelationshipMemberships) {
						rejected = append(rejected, candidate.Tokens)
						mergeCandidates = append(mergeCandidates[:i], mergeCandidates[i+1:]...)
					} else {
						var isNew = true
						for _, annoNameOccurrence := range candidate.annotationNameOccurrences {
							if annoNameOccurrence == relationship.AnnotationName {
								isNew = false
							}
						}
						if isNew {
							mergeCandidates[i].annotationNameOccurrences = append(mergeCandidates[i].annotationNameOccurrences, relationship.AnnotationName)
						}
					}
				} else {
					mergeCandidates, rejected = testRelationshipRejection(relationship, mergeCandidates, rejected)
				}
			}
		}

	}
	return setRelationshipMergeStatus(relationshipAlternatives, mergeCandidates, numberOfAnnotations)
}

func setToreCodeMergeStatus(
	toreAlternatives []TORECodeAlternatives,
	mergeCandidates []ToreCodeMergeCandidate,
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

func setWordCodeMergeStatus(
	wordCodeAlternatives []WordCodeAlternatives,
	mergeCandidates []WordCodeMergeCandidate,
	numberOfAnnotations int,
) []WordCodeAlternatives {

	for _, candidate := range mergeCandidates {
		if len(candidate.annotationNameOccurrences) == numberOfAnnotations {
			var isAccepted = false
			for i, tore := range wordCodeAlternatives {
				if !isAccepted {
					if testEqSlice(candidate.Tokens, tore.Tokens) {
						wordCodeAlternatives[i].MergeStatus = "Accepted"
						isAccepted = true
					}
				} else {
					if testEqSlice(candidate.Tokens, tore.Tokens) {
						wordCodeAlternatives[i].MergeStatus = "Declined"
					}
				}
			}
		}
	}
	return wordCodeAlternatives
}

func setRelationshipMergeStatus(
	relationshipAlternatives []RelationshipAlternatives,
	mergeCandidates []RelationshipMergeCandidate,
	numberOfAnnotations int,
) []RelationshipAlternatives {

	for _, candidate := range mergeCandidates {
		if len(candidate.annotationNameOccurrences) == numberOfAnnotations {
			var isAccepted = false
			for i, tore := range relationshipAlternatives {
				if !isAccepted {
					if testEqSlice(candidate.Tokens, tore.Tokens) {
						relationshipAlternatives[i].MergeStatus = "Accepted"
						isAccepted = true
					}
				} else {
					if testEqSlice(candidate.Tokens, tore.Tokens) {
						relationshipAlternatives[i].MergeStatus = "Declined"
					}
				}
			}
		}
	}
	return relationshipAlternatives
}

func testToreCodeRejection(
	tore TORECodeAlternatives,
	mergeCandidates []ToreCodeMergeCandidate,
	rejected [][]*int,
) ([]ToreCodeMergeCandidate, [][]*int) {

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
		mergeCandidates = append(mergeCandidates, ToreCodeMergeCandidate{tore.Tokens, tore.Tore, []string{tore.AnnotationName}})
	}
	return mergeCandidates, rejected
}

func testWordCodeRejection(
	wordCode WordCodeAlternatives,
	mergeCandidates []WordCodeMergeCandidate,
	rejected [][]*int,
) ([]WordCodeMergeCandidate, [][]*int) {

	var isAReject = false
	for _, reject := range rejected {
		if testEqSlice(wordCode.Tokens, reject) {
			isAReject = true
			rejected = append(rejected, wordCode.Tokens)
			for i, candidate := range mergeCandidates {
				if testEqSlice(wordCode.Tokens, candidate.Tokens) {
					mergeCandidates = append(mergeCandidates[:i], mergeCandidates[i+1:]...)
				}
			}
		}
	}
	if !isAReject {
		mergeCandidates = append(mergeCandidates, WordCodeMergeCandidate{wordCode.Tokens, wordCode.Name, []string{wordCode.AnnotationName}})
	}
	return mergeCandidates, rejected
}

func testRelationshipRejection(
	relationship RelationshipAlternatives,
	mergeCandidates []RelationshipMergeCandidate,
	rejected [][]*int,
) ([]RelationshipMergeCandidate, [][]*int) {

	var isAReject = false
	for _, reject := range rejected {
		if testEqSlice(relationship.Tokens, reject) {
			isAReject = true
			rejected = append(rejected, relationship.Tokens)
			for i, candidate := range mergeCandidates {
				if testEqSlice(relationship.Tokens, candidate.Tokens) {
					mergeCandidates = append(mergeCandidates[:i], mergeCandidates[i+1:]...)
				}
			}
		}
	}
	if !isAReject {
		mergeCandidates = append(mergeCandidates, RelationshipMergeCandidate{relationship.Tokens, relationship.RelationshipMemberships, []string{relationship.AnnotationName}})
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
