package ml

import "testing"

func TestProbabilityToBetSizeIsCappedAndOneSided(t *testing.T) {
	if got := ProbabilityToBetSize(0.49, 0.20, 2); got != 0 {
		t.Fatalf("size below edge=%f, want 0", got)
	}
	if got := ProbabilityToBetSize(0.60, 0.05, 2); got != 0.05 {
		t.Fatalf("capped size=%f, want 0.05", got)
	}
}

func TestProbabilityToBetSizeUsesConfiguredThreshold(t *testing.T) {
	if got := ProbabilityToBetSizeAboveThreshold(0.44, 0.45, 0.20, 2); got != 0 {
		t.Fatalf("size below configured threshold=%f, want 0", got)
	}
	if got := ProbabilityToBetSizeAboveThreshold(0.46, 0.45, 0.20, 2); got <= 0 {
		t.Fatalf("size above configured threshold=%f, want positive", got)
	}
}

func TestValidateParityReportsToleranceFailure(t *testing.T) {
	maxErr, err := ValidateParity([]float64{0.1, 0.2}, []float64{0.1, 0.200001}, 1e-8)
	if err == nil {
		t.Fatalf("expected parity failure")
	}
	if maxErr <= 0 {
		t.Fatalf("max error=%f, want positive", maxErr)
	}
}
