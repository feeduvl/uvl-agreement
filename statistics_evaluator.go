package main

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

	// Get and count all wordCodes
	var nameSet = map[string]bool{}
	var annotationSet = map[string]bool{}
	for _, alternative := range agreement.CodeAlternatives {
		nameSet[alternative.Code.Name] = true
		annotationSet[alternative.AnnotationName] = true
	}

	numberOfRels, numberOfCategories, numberOfWordCodes := getNumberOfRelsCategoriesAndWordCodes(nameSet, toreRelationships, toreCategories)

	var numberOfCategoryAlternatives = numberOfRels * numberOfCategories * numberOfWordCodes

	// Get and encode relationships in agreement
	var existingRelsMap = map[int]string{}
	for _, existingToreRel := range agreement.TORERelationships {
		if existingToreRel.Index != nil {
			existingRelsMap[*existingToreRel.Index] = existingToreRel.RelationshipName
		}
	}

	// Maps to look up position of code in dataMatrix
	relNameMap, categoryMap, wordCodeMap := createMaps(agreement, toreCategories, toreRelationships)

	// Datamatrix containing codes for all tokens
	dataMatrix, dataMatrixForRowCalculation, sumOfAllCells, unmberOfAssignedTokens := fillDataMatrices(agreement.CodeAlternatives, agreement, wordCodeMap, categoryMap, relNameMap, existingRelsMap, numberOfCategories, numberOfRels, len(annotationSet), numberOfCategoryAlternatives)
	fleissKappa, brennanKappa := calculateKappas(numberOfCategoryAlternatives, dataMatrix, dataMatrixForRowCalculation, sumOfAllCells, unmberOfAssignedTokens)
	if fleissKappa < 0 {
		fleissKappa = 0
	}
	if brennanKappa < 0 {
		brennanKappa = 0
	}
	return fleissKappa, brennanKappa
}

func calculateKappas(
	numberOfAlternatives int,
	dataMatrix [][]int,
	dataMatrixForRowCalculation [][]int,
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
			sumOfCodesInRow += float64(dataMatrixForRowCalculation[i][j])
			addedSquaresOfRow += float64(dataMatrixForRowCalculation[i][j] * dataMatrixForRowCalculation[i][j])
		}
		pi[i] = (addedSquaresOfRow - sumOfCodesInRow) / (sumOfCodesInRow * (sumOfCodesInRow - 1))
		sumOfPi += pi[i]
	}
	var pHead = sumOfPi / float64(numberOfTokens)

	var fleissKappa float64
	// This is only for the special case, when denominator is 0
	if (1.0 - pc) == 0.0 {
		fleissKappa = 1.0
	} else {
		fleissKappa = (pHead - pc) / (1.0 - pc)
	}

	// Calculate Brennan and Prediger Kappa
	var brennanPc = sumOfAllCells / ((sumOfAllCells + 1) * (sumOfAllCells + 1))
	var brennanKappa = (pHead - brennanPc) / (1 - brennanPc)
	return fleissKappa, brennanKappa
}

func calculatePosition(
	codeAlternative CodeAlternatives,
	wordCodeMap map[string]int,
	categoryMap map[string]int,
	relNameMap map[string]int,
	existingRelsMap map[int]string,
	numberOfCategories int,
	numberOfRels int,
	dataMatrix [][]int,
	sumOfAllCells int,
	dataRow int,
) (int, int) {
	var totalPosition = 0
	if len(codeAlternative.Code.RelationshipMemberships) != 0 {
		for _, memberIndex := range codeAlternative.Code.RelationshipMemberships {
			var categoryPosition = categoryMap[codeAlternative.Code.Tore]
			var wordCodePosition = wordCodeMap[codeAlternative.Code.Name]
			var relationshipName = existingRelsMap[*memberIndex]
			var relationshipPosition = relNameMap[relationshipName]
			totalPosition = (wordCodePosition * numberOfCategories * numberOfRels) + (categoryPosition * numberOfRels) + relationshipPosition
			dataMatrix[dataRow][totalPosition] += 1
			sumOfAllCells++
		}
	} else {
		var categoryPosition = categoryMap[codeAlternative.Code.Tore]
		var wordCodePosition = wordCodeMap[codeAlternative.Code.Name]
		var relationshipPosition = 0
		totalPosition = (wordCodePosition * numberOfCategories * numberOfRels) + (categoryPosition * numberOfRels) + relationshipPosition
		dataMatrix[dataRow][totalPosition] += 1
		sumOfAllCells++
	}
	return totalPosition, sumOfAllCells
}

func fillDataMatrices(
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
) ([][]int, [][]int, float64, int) {
	var tokenMap = map[int][]CodeAlternatives{}
	var sumOfAllCells = 0

	for _, codeAlternative := range codeAlternatives {
		// Ignore declined
		if codeAlternative.MergeStatus != "Declined" {
			for _, token := range codeAlternative.Code.Tokens {
				tokenMap[*token] = append(tokenMap[*token], codeAlternative)
			}
		}
	}

	var dataMatrix = make([][]int, len(tokenMap))
	var dataMatrixForRowCalculations = make([][]int, len(tokenMap))
	for i := 0; i < len(tokenMap); i++ {
		// looping through the slice to declare
		// slice of slice of correct length
		dataMatrix[i] = make([]int, numberOfAlternatives)
		dataMatrixForRowCalculations[i] = make([]int, numberOfAlternatives)
	}
	dataRow := 0
	for _, token := range agreement.Tokens {
		var tokenIndex = *token.Index
		if len(tokenMap[tokenIndex]) == 0 {
			continue
		}
		var acceptedFieldToFill = 0
		var isFirstAccepted = true
		var annotationNameSet = map[string]bool{}
		for _, codeAlternative := range tokenMap[tokenIndex] {
			_, sumOfAllCells = calculatePosition(codeAlternative, wordCodeMap, categoryMap, relNameMap, existingRelsMap, numberOfCategories, numberOfRels, dataMatrix, sumOfAllCells, dataRow)
			if codeAlternative.MergeStatus == "Accepted" {
				if isFirstAccepted {
					var totalPosition int
					totalPosition, _ = calculatePosition(codeAlternative, wordCodeMap, categoryMap, relNameMap, existingRelsMap, numberOfCategories, numberOfRels, dataMatrixForRowCalculations, sumOfAllCells, dataRow)
					acceptedFieldToFill = totalPosition
					isFirstAccepted = false
				} else {
					dataMatrixForRowCalculations[dataRow][acceptedFieldToFill] += 1
				}
			} else {
				calculatePosition(codeAlternative, wordCodeMap, categoryMap, relNameMap, existingRelsMap, numberOfCategories, numberOfRels, dataMatrixForRowCalculations, sumOfAllCells, dataRow)
			}
			annotationNameSet[codeAlternative.AnnotationName] = true
		}
		var numberOfUnassignedCodes = numberOfAnnotations - len(annotationNameSet)
		if numberOfUnassignedCodes > 0 {
			dataMatrix[dataRow][acceptedFieldToFill] += numberOfUnassignedCodes
			dataMatrixForRowCalculations[dataRow][acceptedFieldToFill] += numberOfUnassignedCodes
			sumOfAllCells += numberOfUnassignedCodes
		}
		dataRow++
	}
	return dataMatrix, dataMatrixForRowCalculations, float64(sumOfAllCells), len(tokenMap)
}

func getNumberOfRelsCategoriesAndWordCodes(
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
