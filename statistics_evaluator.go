package main

import "fmt"

type ToreCategories struct {
	Tores []string `json:"tores" bson:"tores"`
}
type ToreRelationships struct {
	Owners            []string `json:"owners" bson:"owners"`
	RelationshipNames []string `json:"relationship_names" bson:"relationship_names"`
}

func getKappas(
	agreement Agreement,
	toreCategories ToreCategories,
	toreRelationships ToreRelationships,
) (float64, float64) {

	var nameSet = map[string]bool{}
	var annotationSet = map[string]bool{}
	for _, alternative := range agreement.CodeAlternatives {
		nameSet[alternative.Code.Name] = true
		annotationSet[alternative.AnnotationName] = true
	}

	numberOfRels, numberOfCategories, numberOfWordCodes := getNumberOfCategories(nameSet, toreRelationships, toreCategories)

	var numberOfAlternatives = numberOfRels * numberOfCategories * numberOfWordCodes

	var existingRelsMap = map[int]string{}
	for _, existingToreRel := range agreement.TORERelationships {
		if existingToreRel.Index != nil {
			existingRelsMap[*existingToreRel.Index] = existingToreRel.RelationshipName
		}
	}

	relNameMap, categoryMap, wordCodeMap := createMaps(agreement, toreCategories, toreRelationships)

	dataMatrix, sumOfAllCells, numberOfAssignedTokens := fillDataMatrix(agreement.CodeAlternatives, agreement, wordCodeMap, categoryMap, relNameMap, existingRelsMap, numberOfCategories, numberOfRels, len(annotationSet), numberOfAlternatives)
	fleissKappa, brennanKappa := calculateKappas(numberOfAlternatives, dataMatrix, sumOfAllCells, numberOfAssignedTokens)
	return fleissKappa, brennanKappa
}

func calculateKappas(
	numberOfAlternatives int,
	dataMatrix [][]int,
	sumOfAllCells float64,
	numberOfTokens int,
) (float64, float64) {
	// Calculate Fleiss Kappa
	var pj = make([]float64, numberOfAlternatives)
	var pc = float64(0)
	for j := 0; j < numberOfAlternatives; j++ {
		var sumOfColumn = 0
		for i, _ := range dataMatrix {
			sumOfColumn += dataMatrix[i][j]
		}
		pj[j] = float64(sumOfColumn) / sumOfAllCells
		pc += pj[j] * pj[j]
	}
	var pi = make([]float64, numberOfTokens)
	var sumOfPi = float64(0)
	for i := 0; i < numberOfTokens; i++ {
		var sumOfCodesInRow = float64(0)
		var addedSquaresOfRow = float64(0)
		for j := 0; j < numberOfAlternatives; j++ {
			sumOfCodesInRow += float64(dataMatrix[i][j])
			addedSquaresOfRow += float64(dataMatrix[i][j] * dataMatrix[i][j])
		}
		pi[i] = (addedSquaresOfRow - sumOfCodesInRow) / (sumOfCodesInRow * (sumOfCodesInRow - 1))
		sumOfPi += pi[i]
	}
	var pHead = sumOfPi / float64(numberOfTokens)
	fmt.Printf("numberOfTokens: %v\n", numberOfTokens)
	fmt.Printf("sumOfPi: %v\n", sumOfPi)

	var fleissKappa float64
	// This is only for the special case, when denominator is 0
	if (1.0 - pc) == 0.0 {
		fleissKappa = 1.0
	} else {
		fleissKappa = (pHead - pc) / (1.0 - pc)
	}

	// Calculate Brennan and Prediger Kappa
	var brennanPc = sumOfAllCells / ((sumOfAllCells + 1) * (sumOfAllCells + 1))
	fmt.Printf("Brennanpc: %v\n", brennanPc)
	fmt.Printf("phead: %v\n", pHead)
	var brennanKappa float64
	if (1 - brennanPc) == 0.0 {
		brennanKappa = 1.0
	} else {
		brennanKappa = (pHead - brennanPc) / (1 - brennanPc)
	}
	return fleissKappa, brennanKappa
}

func fillDataMatrix(
	codeAlternatives []CodeAlternatives,
	agreement Agreement,
	wordCodeMap map[string]int,
	categoryMap map[string]int,
	relNameMap map[string]int,
	existingRelsMap map[int]string,
	numberOfCategories int,
	numberOfRels int,
	numberOfAnnotations int,
	numberOfAlternatives int,
) ([][]int, float64, int) {
	var tokenMap = map[int][]CodeAlternatives{}
	var sumOfAllCells = 0

	for _, codeAlternative := range codeAlternatives {
		for _, token := range codeAlternative.Code.Tokens {
			tokenMap[*token] = append(tokenMap[*token], codeAlternative)
		}
	}

	var dataMatrix = make([][]int, len(tokenMap))
	for i := 0; i < len(tokenMap); i++ {
		// looping through the slice to declare
		// slice of slice of correct length
		dataMatrix[i] = make([]int, numberOfAlternatives)
	}
	dataRow := 0
	for _, token := range agreement.Tokens {
		var tokenIndex = *token.Index
		if len(tokenMap[tokenIndex]) == 0 {
			continue
		}
		var annotationNameSet = map[string]bool{}
		for _, codeAlternative := range tokenMap[tokenIndex] {
			annotationNameSet[codeAlternative.AnnotationName] = true
			if len(codeAlternative.Code.RelationshipMemberships) != 0 {
				for _, memberIndex := range codeAlternative.Code.RelationshipMemberships {
					var categoryPosition = categoryMap[codeAlternative.Code.Tore]
					var wordCodePosition = wordCodeMap[codeAlternative.Code.Name]
					var relationshipName = existingRelsMap[*memberIndex]
					var relationshipPosition = relNameMap[relationshipName]
					var totalPosition = (wordCodePosition * numberOfCategories * numberOfRels) + (categoryPosition * numberOfRels) + relationshipPosition
					dataMatrix[dataRow][totalPosition] += 1
					sumOfAllCells++
				}
			} else {
				var categoryPosition = categoryMap[codeAlternative.Code.Tore]
				var wordCodePosition = wordCodeMap[codeAlternative.Code.Name]
				var relationshipPosition = 0
				var totalPosition = (wordCodePosition * numberOfCategories * numberOfRels) + (categoryPosition * numberOfRels) + relationshipPosition
				dataMatrix[dataRow][totalPosition] += 1
				sumOfAllCells++
			}
		}
		var numberOfUnassignedCodes = numberOfAnnotations - len(annotationNameSet)
		dataMatrix[dataRow][0] = numberOfUnassignedCodes
		sumOfAllCells += numberOfUnassignedCodes
		dataRow++
	}
	return dataMatrix, float64(sumOfAllCells), len(tokenMap)
}

func getNumberOfCategories(
	nameSet map[string]bool,
	toreRelationships ToreRelationships,
	toreCategories ToreCategories,
) (int, int, int) {

	var numberOfRels = len(toreRelationships.RelationshipNames) + 1
	var numberOfCategories = len(toreCategories.Tores) + 1
	var numberOfWordCodes = len(nameSet)
	if _, ok := nameSet[""]; !ok {
		numberOfWordCodes++
	}

	return numberOfRels, numberOfCategories, numberOfWordCodes
}

func createMaps(
	agreement Agreement,
	toreCategories ToreCategories,
	toreRelationships ToreRelationships,
) (
	map[string]int,
	map[string]int,
	map[string]int,
) {
	var relNameMap = map[string]int{}
	var categoryMap = map[string]int{}
	var wordCodeMap = map[string]int{}

	var wordMapCounter = 1
	var relMapCounter = 1
	var catMapCounter = 1
	relNameMap[""] = 0
	categoryMap[""] = 0
	wordCodeMap[""] = 0

	for _, alternative := range agreement.CodeAlternatives {
		if _, ok := wordCodeMap[alternative.Code.Name]; !ok {
			wordCodeMap[alternative.Code.Name] = wordMapCounter
			wordMapCounter++
		}
	}

	for _, toreCategory := range toreCategories.Tores {
		if _, ok := categoryMap[toreCategory]; !ok {
			categoryMap[toreCategory] = catMapCounter
			catMapCounter++
		}
	}
	for _, toreRelName := range toreRelationships.RelationshipNames {
		if _, ok := relNameMap[toreRelName]; !ok {
			relNameMap[toreRelName] = relMapCounter
			relMapCounter++
		}
	}

	return relNameMap, categoryMap, wordCodeMap
}
