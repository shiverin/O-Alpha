package validation

import (
	"testing"

	"github.com/oalpha/internal/backtest"
)

func TestEvaluatePromotionPromotesRiskAdjustedImprovement(t *testing.T) {
	metrics := backtest.PortfolioMetrics{
		AnnualReturn: 0.11,
		Sortino:      1.4,
		Calmar:       1.2,
		MaxDrawdown:  0.08,
		NumTrades:    40,
		Turnover:     1.0,
		DSR:          0.98,
		PBO:          0.10,
	}
	benchmark := backtest.PortfolioMetrics{
		AnnualReturn: 0.10,
		Sortino:      1.0,
		Calmar:       0.8,
		MaxDrawdown:  0.12,
		NumTrades:    40,
		Turnover:     1.0,
	}

	decision := EvaluatePromotion(metrics, benchmark, DefaultPromotionConfig(), true, true)
	if !decision.Promote {
		t.Fatalf("expected promotion, reasons=%v", decision.Reasons)
	}
}

func TestEvaluatePromotionRejectsFailedGovernance(t *testing.T) {
	metrics := backtest.PortfolioMetrics{
		AnnualReturn: 0.11,
		Sortino:      1.4,
		Calmar:       1.2,
		MaxDrawdown:  0.08,
		NumTrades:    5,
		Turnover:     2.0,
		DSR:          0.5,
		PBO:          0.5,
	}
	benchmark := backtest.PortfolioMetrics{
		AnnualReturn: 0.10,
		Sortino:      1.0,
		Calmar:       0.8,
		MaxDrawdown:  0.12,
		NumTrades:    40,
		Turnover:     1.0,
	}

	decision := EvaluatePromotion(metrics, benchmark, DefaultPromotionConfig(), false, false)
	if decision.Promote {
		t.Fatalf("expected rejection")
	}
	if len(decision.Reasons) < 4 {
		t.Fatalf("expected multiple rejection reasons, got %v", decision.Reasons)
	}
}

func TestProbabilityOfBacktestOverfitting(t *testing.T) {
	pbo, err := ProbabilityOfBacktestOverfitting(
		[][]float64{
			{1.0, 0.5, 0.2},
			{0.4, 1.2, 0.1},
		},
		[][]float64{
			{0.1, 0.4, 0.5},
			{0.5, 0.1, 0.4},
		},
	)
	if err != nil {
		t.Fatalf("ProbabilityOfBacktestOverfitting: %v", err)
	}
	if pbo != 1 {
		t.Fatalf("pbo=%f, want 1 for train winners underperforming OOS", pbo)
	}
}

func TestDeflatedSharpeRatioIncreasesWithSharpe(t *testing.T) {
	low := DeflatedSharpeRatio(0.1, 252, 0, 3, 10)
	high := DeflatedSharpeRatio(1.0, 252, 0, 3, 10)
	if high <= low {
		t.Fatalf("high DSR=%f should exceed low DSR=%f", high, low)
	}
}
