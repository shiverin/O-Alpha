package validation

import "fmt"

type MLWalkForwardConfig struct {
	TrainSize      int `json:"train_size"`
	ValidationSize int `json:"validation_size"`
	TestSize       int `json:"test_size"`
	StepSize       int `json:"step_size"`
	EmbargoBars    int `json:"embargo_bars"`
}

type WalkForwardSplit struct {
	Fold            int `json:"fold"`
	TrainStart      int `json:"train_start"`
	TrainEnd        int `json:"train_end"`
	ValidationStart int `json:"validation_start"`
	ValidationEnd   int `json:"validation_end"`
	TestStart       int `json:"test_start"`
	TestEnd         int `json:"test_end"`
}

func GenerateMLWalkForwardSplits(numSamples int, cfg MLWalkForwardConfig) ([]WalkForwardSplit, error) {
	if numSamples <= 0 {
		return nil, fmt.Errorf("walk-forward split requires samples")
	}
	if cfg.TrainSize <= 0 {
		return nil, fmt.Errorf("train_size must be positive")
	}
	if cfg.TestSize <= 0 {
		return nil, fmt.Errorf("test_size must be positive")
	}
	if cfg.StepSize <= 0 {
		cfg.StepSize = cfg.TestSize
	}
	if cfg.ValidationSize < 0 {
		cfg.ValidationSize = 0
	}
	if cfg.EmbargoBars < 0 {
		cfg.EmbargoBars = 0
	}

	splits := make([]WalkForwardSplit, 0)
	fold := 0
	for trainStart := 0; ; trainStart += cfg.StepSize {
		trainEnd := trainStart + cfg.TrainSize
		validationStart := trainEnd
		validationEnd := validationStart + cfg.ValidationSize
		testStart := validationEnd + cfg.EmbargoBars
		testEnd := testStart + cfg.TestSize
		if testEnd > numSamples {
			break
		}
		splits = append(splits, WalkForwardSplit{
			Fold:            fold,
			TrainStart:      trainStart,
			TrainEnd:        trainEnd,
			ValidationStart: validationStart,
			ValidationEnd:   validationEnd,
			TestStart:       testStart,
			TestEnd:         testEnd,
		})
		fold++
	}
	if len(splits) == 0 {
		return nil, fmt.Errorf("configuration produces no walk-forward splits for %d samples", numSamples)
	}
	return splits, nil
}
