package alphavalidation

import (
	"fmt"
	"math"
	"sort"
)

type scoredFold struct {
	variant string
	score   float64
}

func EstimatePBO(foldScores map[string][]float64, objective string) PBODiagnostics {
	if len(foldScores) < 3 {
		return PBODiagnostics{
			Estimated:    false,
			Method:       "train-half rank inversion proxy",
			Objective:    objective,
			VariantCount: len(foldScores),
			FailureReason: "need at least 3 real sibling variants",
		}
	}

	foldCount := -1
	for _, scores := range foldScores {
		if foldCount == -1 {
			foldCount = len(scores)
			continue
		}
		if len(scores) != foldCount {
			return PBODiagnostics{Estimated: false, Method: "train-half rank inversion proxy", Objective: objective, VariantCount: len(foldScores), FailureReason: "variant fold counts are inconsistent"}
		}
	}
	if foldCount < 2 {
		return PBODiagnostics{Estimated: false, Method: "train-half rank inversion proxy", Objective: objective, VariantCount: len(foldScores), FoldCount: foldCount, FailureReason: "need at least 2 folds for PBO estimation"}
	}

	trainIdx := make([]int, 0, (foldCount+1)/2)
	testIdx := make([]int, 0, foldCount/2)
	for i := 0; i < foldCount; i++ {
		if i%2 == 0 {
			trainIdx = append(trainIdx, i)
		} else {
			testIdx = append(testIdx, i)
		}
	}
	if len(testIdx) == 0 {
		testIdx = append(testIdx, trainIdx[len(trainIdx)-1])
		trainIdx = trainIdx[:len(trainIdx)-1]
	}
	if len(trainIdx) == 0 || len(testIdx) == 0 {
		return PBODiagnostics{Estimated: false, Method: "train-half rank inversion proxy", Objective: objective, VariantCount: len(foldScores), FoldCount: foldCount, FailureReason: "unable to split folds into train/test halves"}
	}

	trainRanks := make([]scoredFold, 0, len(foldScores))
	testRanks := make([]scoredFold, 0, len(foldScores))
	for variant, scores := range foldScores {
		trainRanks = append(trainRanks, scoredFold{variant: variant, score: averageAt(scores, trainIdx)})
		testRanks = append(testRanks, scoredFold{variant: variant, score: averageAt(scores, testIdx)})
	}
	sort.Slice(trainRanks, func(i, j int) bool { return trainRanks[i].score > trainRanks[j].score })
	sort.Slice(testRanks, func(i, j int) bool { return testRanks[i].score > testRanks[j].score })

	trainWinner := trainRanks[0].variant
	testRank := 0
	for i, ranked := range testRanks {
		if ranked.variant == trainWinner {
			testRank = i + 1
			break
		}
	}
	if testRank == 0 {
		return PBODiagnostics{Estimated: false, Method: "train-half rank inversion proxy", Objective: objective, VariantCount: len(foldScores), FoldCount: foldCount, FailureReason: fmt.Sprintf("train winner %s not found in test ranks", trainWinner)}
	}

	// Convert 1-based rank (1 = best) into a performance percentile where higher is better.
	performancePct := 1.0 - (float64(testRank) / float64(len(testRanks)+1))
	if performancePct <= 0 {
		performancePct = 1e-9
	}
	if performancePct >= 1 {
		performancePct = 1 - 1e-9
	}
	lambda := math.Log(performancePct / (1 - performancePct))
	pbo := 0.0
	if lambda < 0 {
		pbo = 1.0
	}

	return PBODiagnostics{
		Estimated:         true,
		Method:            "train-half rank inversion proxy",
		Objective:         objective,
		VariantCount:      len(foldScores),
		FoldCount:         foldCount,
		SplitCount:        1,
		PBO:               pbo,
		MedianLambda:      lambda,
		TrainWinnerCounts: map[string]int{trainWinner: 1},
	}
}

func averageAt(scores []float64, indexes []int) float64 {
	if len(indexes) == 0 {
		return 0
	}
	total := 0.0
	count := 0
	for _, idx := range indexes {
		if idx < 0 || idx >= len(scores) {
			continue
		}
		total += scores[idx]
		count++
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}
