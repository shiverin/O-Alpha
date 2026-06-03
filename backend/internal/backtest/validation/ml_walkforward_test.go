package validation

import "testing"

func TestGenerateMLWalkForwardSplits(t *testing.T) {
	splits, err := GenerateMLWalkForwardSplits(30, MLWalkForwardConfig{
		TrainSize:      10,
		ValidationSize: 2,
		TestSize:       5,
		StepSize:       5,
		EmbargoBars:    1,
	})
	if err != nil {
		t.Fatalf("GenerateMLWalkForwardSplits: %v", err)
	}
	if len(splits) != 3 {
		t.Fatalf("splits=%d, want 3", len(splits))
	}
	first := splits[0]
	if first.TrainStart != 0 || first.TrainEnd != 10 || first.ValidationStart != 10 || first.ValidationEnd != 12 || first.TestStart != 13 || first.TestEnd != 18 {
		t.Fatalf("unexpected first split: %+v", first)
	}
}
