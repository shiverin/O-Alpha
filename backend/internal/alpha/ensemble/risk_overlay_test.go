package ensemble

import (
	"math"
	"slices"
	"testing"

	"github.com/oalpha/internal/backtest"
)

func TestApplyPortfolioRiskOverlayScalesHighRealizedVol(t *testing.T) {
	targets := map[string]backtest.TargetPosition{
		"AAPL": {Symbol: "AAPL", TargetWeight: 0.50},
	}
	returns := make([]float64, 80)
	for i := range returns {
		if i%2 == 0 {
			returns[i] = 0.05
		} else {
			returns[i] = -0.05
		}
	}
	result := ApplyPortfolioRiskOverlay(targets, PortfolioRiskConfig{
		TargetVolAnnual:  0.10,
		VolLookbackDays:  60,
		MaxSymbolWeight:  1,
		MaxGrossExposure: 1,
		MaxNetExposure:   1,
		FloorVolAnnual:   0.01,
	}, PortfolioRiskState{PortfolioReturns: returns}, nil)

	if result.RiskScalar >= 1 {
		t.Fatalf("risk scalar=%f, want reduction", result.RiskScalar)
	}
	if result.Targets["AAPL"].TargetWeight >= 0.50 {
		t.Fatalf("target not reduced: %f", result.Targets["AAPL"].TargetWeight)
	}
	if !slices.Contains(result.Reasons, "realized_vol_target") {
		t.Fatalf("reasons=%v, want realized_vol_target", result.Reasons)
	}
}

func TestApplyPortfolioRiskOverlayHardDrawdownStopFlattens(t *testing.T) {
	targets := map[string]backtest.TargetPosition{
		"AAPL": {Symbol: "AAPL", TargetWeight: 0.50},
	}
	result := ApplyPortfolioRiskOverlay(targets, PortfolioRiskConfig{
		TargetVolAnnual:     1,
		MaxSymbolWeight:     1,
		MaxGrossExposure:    1,
		MaxNetExposure:      1,
		MaxDrawdownHardStop: 0.20,
	}, PortfolioRiskState{PeakEquity: 100, CurrentEquity: 70}, nil)

	if result.RiskScalar != 0 {
		t.Fatalf("risk scalar=%f, want hard stop zero", result.RiskScalar)
	}
	if math.Abs(result.Targets["AAPL"].TargetWeight) > 1e-12 {
		t.Fatalf("target=%f, want flat", result.Targets["AAPL"].TargetWeight)
	}
}

func TestApplyPortfolioRiskOverlayCapsSymbolSectorGrossAndNet(t *testing.T) {
	targets := map[string]backtest.TargetPosition{
		"AAPL": {Symbol: "AAPL", TargetWeight: 0.50},
		"MSFT": {Symbol: "MSFT", TargetWeight: 0.50},
		"XOM":  {Symbol: "XOM", TargetWeight: 0.50},
	}
	result := ApplyPortfolioRiskOverlay(targets, PortfolioRiskConfig{
		TargetVolAnnual:  1,
		MaxSymbolWeight:  0.30,
		MaxSectorWeight:  0.40,
		MaxGrossExposure: 0.70,
		MaxNetExposure:   0.60,
	}, PortfolioRiskState{}, map[string]string{
		"AAPL": "tech",
		"MSFT": "tech",
		"XOM":  "energy",
	})

	if result.GrossExposure > 0.70+1e-9 {
		t.Fatalf("gross=%f exceeds cap", result.GrossExposure)
	}
	if math.Abs(result.NetExposure) > 0.60+1e-9 {
		t.Fatalf("net=%f exceeds cap", result.NetExposure)
	}
	if result.Targets["AAPL"].TargetWeight > 0.30+1e-9 {
		t.Fatalf("symbol cap failed: %f", result.Targets["AAPL"].TargetWeight)
	}
}
