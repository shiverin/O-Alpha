package alphavalidation

import (
	"fmt"
	"time"
)

type WalkForwardWindow struct {
	Fold       int
	TrainStart int
	TrainEnd   int
	TestStart  int
	TestEnd    int
}

func BuildWalkForwardWindows(barCount, trainBars, testBars, stepBars int) ([]WalkForwardWindow, error) {
	if trainBars <= 0 || testBars <= 0 {
		return nil, fmt.Errorf("trainBars and testBars must be positive")
	}
	if stepBars <= 0 {
		stepBars = testBars
	}
	if barCount < trainBars+testBars {
		return nil, fmt.Errorf("insufficient bars for walk-forward: have %d, need %d", barCount, trainBars+testBars)
	}
	windows := make([]WalkForwardWindow, 0)
	fold := 0
	for start := 0; start+trainBars+testBars <= barCount; start += stepBars {
		trainEnd := start + trainBars
		testEnd := trainEnd + testBars
		windows = append(windows, WalkForwardWindow{
			Fold:       fold,
			TrainStart: start,
			TrainEnd:   trainEnd,
			TestStart:  trainEnd,
			TestEnd:    testEnd,
		})
		fold++
	}
	return windows, nil
}

func WindowFromBounds(from, to string) (ValidationWindow, error) {
	start, err := time.Parse("2006-01-02", from)
	if err != nil {
		return ValidationWindow{}, fmt.Errorf("parse from: %w", err)
	}
	end, err := time.Parse("2006-01-02", to)
	if err != nil {
		return ValidationWindow{}, fmt.Errorf("parse to: %w", err)
	}
	start = start.UTC()
	end = end.UTC().Add(24*time.Hour - time.Nanosecond)
	if !end.After(start) {
		return ValidationWindow{}, fmt.Errorf("window end must be after start")
	}
	return ValidationWindow{From: start, To: end}, nil
}
