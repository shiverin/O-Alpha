package alphavalidation

import (
	"testing"

	"github.com/oalpha/pkg/models"
)

func TestEvaluatePromotionFailsWithoutPBO(t *testing.T) {
	candidate := &CandidateReport{Result: &models.BacktestResult{Sharpe: 1.1, Sortino: 1.2, Calmar: 1.3, MaxDrawdown: 0.10, NumTrades: 40}}
	buyHold := &models.BacktestResult{Sharpe: 1.0, Sortino: 1.0, Calmar: 1.0, MaxDrawdown: 0.12}
	decision := EvaluatePromotion(candidate, buyHold, 30)
	if decision.Promote {
		t.Fatalf("expected promotion to fail without PBO")
	}
}

func TestEvaluatePromotionPassesWhenMetricsAndPBOPass(t *testing.T) {
	candidate := &CandidateReport{
		Result: &models.BacktestResult{Sharpe: 1.3, Sortino: 1.5, Calmar: 1.4, MaxDrawdown: 0.10, NumTrades: 40},
		PBO:    &PBODiagnostics{Estimated: true, PBO: 0.10},
	}
	buyHold := &models.BacktestResult{Sharpe: 1.0, Sortino: 1.0, Calmar: 1.0, MaxDrawdown: 0.12}
	decision := EvaluatePromotion(candidate, buyHold, 30)
	if !decision.Promote {
		t.Fatalf("expected promotion to pass, got %s", decision.Reason)
	}
}
