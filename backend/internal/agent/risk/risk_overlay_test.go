package risk

import "testing"

func TestRiskOverlayReducesExposureWithoutCreatingAlpha(t *testing.T) {
	policy := DefaultRiskOverlayPolicy()
	policy.MinConsecutiveBarsToSwitch = 1
	overlay := NewRegimeRiskOverlay(policy)
	decision := overlay.Apply(RegimeOverlayInput{
		BaseExposure:   0.10,
		PosteriorProbs: []float64{0.10, 0.20, 0.70},
		StateRoles:     []RegimeRiskRole{RegimeRiskLowVol, RegimeRiskNormal, RegimeRiskHighVol},
		ModelHealthy:   true,
	})

	if decision.AdjustedExposure >= decision.BaseExposure {
		t.Fatalf("expected high-vol overlay to reduce exposure, got base %.4f adjusted %.4f", decision.BaseExposure, decision.AdjustedExposure)
	}
	if decision.AdjustedExposure <= 0 {
		t.Fatalf("high-vol should size down, not fully veto by default")
	}

	flat := overlay.Apply(RegimeOverlayInput{
		BaseExposure:   0,
		PosteriorProbs: []float64{0.05, 0.15, 0.80},
		StateRoles:     []RegimeRiskRole{RegimeRiskLowVol, RegimeRiskNormal, RegimeRiskHighVol},
		ModelHealthy:   true,
	})
	if flat.AdjustedExposure != 0 {
		t.Fatalf("overlay must not create exposure from flat base alpha, got %.4f", flat.AdjustedExposure)
	}
}

func TestRiskOverlayImmediateCrisisDerisk(t *testing.T) {
	policy := DefaultRiskOverlayPolicy()
	overlay := NewRegimeRiskOverlay(policy)
	decision := overlay.Apply(RegimeOverlayInput{
		BaseExposure:   0.20,
		PosteriorProbs: []float64{0.05, 0.10, 0.05, 0.80},
		StateRoles:     []RegimeRiskRole{RegimeRiskLowVol, RegimeRiskNormal, RegimeRiskHighVol, RegimeRiskCrisis},
		ModelHealthy:   true,
	})

	if decision.EffectiveRole != RegimeRiskCrisis {
		t.Fatalf("expected crisis role, got %s", decision.EffectiveRole)
	}
	if decision.Multiplier != policy.StateMultipliers[RegimeRiskCrisis] {
		t.Fatalf("expected crisis multiplier %.2f, got %.2f", policy.StateMultipliers[RegimeRiskCrisis], decision.Multiplier)
	}
}

func TestRiskOverlayHysteresisRequiresConsecutiveBars(t *testing.T) {
	policy := DefaultRiskOverlayPolicy()
	policy.ImmediateCrisisDeRisk = false
	policy.MinConsecutiveBarsToSwitch = 2
	overlay := NewRegimeRiskOverlay(policy)

	first := overlay.Apply(RegimeOverlayInput{
		BaseExposure:   0.10,
		PosteriorProbs: []float64{0.05, 0.10, 0.85},
		StateRoles:     []RegimeRiskRole{RegimeRiskLowVol, RegimeRiskNormal, RegimeRiskHighVol},
		ModelHealthy:   true,
	})
	if first.EffectiveRole != RegimeRiskNormal {
		t.Fatalf("first high-vol bar should remain normal due hysteresis, got %s", first.EffectiveRole)
	}

	second := overlay.Apply(RegimeOverlayInput{
		BaseExposure:   0.10,
		PosteriorProbs: []float64{0.05, 0.10, 0.85},
		StateRoles:     []RegimeRiskRole{RegimeRiskLowVol, RegimeRiskNormal, RegimeRiskHighVol},
		ModelHealthy:   true,
	})
	if second.EffectiveRole != RegimeRiskHighVol {
		t.Fatalf("second high-vol bar should switch, got %s", second.EffectiveRole)
	}
}

func TestRiskOverlayUnhealthyFailsOpenByDefault(t *testing.T) {
	overlay := NewRegimeRiskOverlay(DefaultRiskOverlayPolicy())
	decision := overlay.Apply(RegimeOverlayInput{
		BaseExposure:   0.10,
		PosteriorProbs: []float64{0.05, 0.10, 0.85},
		StateRoles:     []RegimeRiskRole{RegimeRiskLowVol, RegimeRiskNormal, RegimeRiskHighVol},
		ModelHealthy:   false,
	})

	if decision.AdjustedExposure != decision.BaseExposure {
		t.Fatalf("default unhealthy model should fail open, got %.4f", decision.AdjustedExposure)
	}
}

func TestRiskOverlayFailClosedWhenConfigured(t *testing.T) {
	policy := DefaultRiskOverlayPolicy()
	policy.FailClosedOnUncertain = true
	overlay := NewRegimeRiskOverlay(policy)
	decision := overlay.Apply(RegimeOverlayInput{
		BaseExposure:   0.10,
		PosteriorProbs: []float64{0.05, 0.10, 0.85},
		StateRoles:     []RegimeRiskRole{RegimeRiskLowVol, RegimeRiskNormal, RegimeRiskHighVol},
		ModelHealthy:   false,
	})

	if decision.AdjustedExposure != 0 {
		t.Fatalf("fail-closed unhealthy model should zero exposure, got %.4f", decision.AdjustedExposure)
	}
	if !decision.Vetoed {
		t.Fatalf("fail-closed zeroing should be marked vetoed")
	}
}

func TestRiskOverlayRealizedVolCap(t *testing.T) {
	policy := DefaultRiskOverlayPolicy()
	policy.MaxRealizedAnnualVol = 0.20
	policy.VolCapMultiplier = 0.30
	overlay := NewRegimeRiskOverlay(policy)
	decision := overlay.Apply(RegimeOverlayInput{
		BaseExposure:      0.10,
		PosteriorProbs:    []float64{0.80, 0.10, 0.10},
		StateRoles:        []RegimeRiskRole{RegimeRiskLowVol, RegimeRiskNormal, RegimeRiskHighVol},
		ModelHealthy:      true,
		RealizedAnnualVol: 0.35,
	})

	if decision.Multiplier != 0.30 {
		t.Fatalf("expected vol cap multiplier 0.30, got %.2f", decision.Multiplier)
	}
}

func TestRiskOverlayDrawdownGuard(t *testing.T) {
	policy := DefaultRiskOverlayPolicy()
	policy.MaxDrawdownPct = 0.05
	policy.DrawdownMultiplier = 0.40
	overlay := NewRegimeRiskOverlay(policy)
	decision := overlay.Apply(RegimeOverlayInput{
		BaseExposure:   0.10,
		PosteriorProbs: []float64{0.80, 0.10, 0.10},
		StateRoles:     []RegimeRiskRole{RegimeRiskLowVol, RegimeRiskNormal, RegimeRiskHighVol},
		ModelHealthy:   true,
		PeakEquity:     100_000,
		CurrentEquity:  90_000,
	})

	if decision.Multiplier != 0.40 {
		t.Fatalf("expected drawdown guard multiplier 0.40, got %.2f", decision.Multiplier)
	}
}
