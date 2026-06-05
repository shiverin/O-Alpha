package alphavalidation

import "testing"

func TestEstimatePBOFailsClosedWithoutEnoughVariants(t *testing.T) {
	diagnostics := EstimatePBO(map[string][]float64{
		"a": {1, 2, 3},
		"b": {2, 3, 4},
	}, "score")
	if diagnostics.Estimated {
		t.Fatalf("expected PBO estimation to fail with fewer than 3 variants")
	}
}

func TestEstimatePBODetectsRankInversion(t *testing.T) {
	diagnostics := EstimatePBO(map[string][]float64{
		"winner": {10, 1, 9, 1},
		"steady": {7, 7, 7, 7},
		"late":   {6, 8, 6, 8},
	}, "score")
	if !diagnostics.Estimated {
		t.Fatalf("expected PBO estimation to succeed: %s", diagnostics.FailureReason)
	}
	if diagnostics.PBO <= 0 {
		t.Fatalf("expected positive PBO for train/test inversion, got %.3f", diagnostics.PBO)
	}
}
