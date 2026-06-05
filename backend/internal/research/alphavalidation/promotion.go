package alphavalidation

import (
	"fmt"

	"github.com/oalpha/pkg/models"
)

const maxAllowedPBO = 0.20

func EvaluatePromotion(candidate *CandidateReport, buyHold *models.BacktestResult, minTrades int) PromotionDecision {
	decision := PromotionDecision{}
	if candidate == nil || candidate.Result == nil {
		decision.Reason = "candidate result is required"
		return decision
	}
	if buyHold == nil {
		decision.Reason = "buy-hold benchmark is required"
		return decision
	}
	if candidate.PBO == nil || !candidate.PBO.Estimated {
		decision.Reason = "PBO not estimated"
		if candidate.PBO != nil {
			decision.PBO = candidate.PBO.PBO
			decision.PBOEstimated = candidate.PBO.Estimated
		}
		return decision
	}

	decision.PBOEstimated = true
	decision.PBO = candidate.PBO.PBO
	decision.SharpeDeltaVsBuyHold = candidate.Result.Sharpe - buyHold.Sharpe
	decision.SortinoDeltaVsBuyHold = candidate.Result.Sortino - buyHold.Sortino
	decision.CalmarDeltaVsBuyHold = candidate.Result.Calmar - buyHold.Calmar
	decision.DrawdownDeltaVsBuyHold = buyHold.MaxDrawdown - candidate.Result.MaxDrawdown
	decision.TradeCountOK = candidate.Result.NumTrades >= minTrades

	switch {
	case candidate.PBO.PBO > maxAllowedPBO:
		decision.Reason = fmt.Sprintf("PBO %.3f above %.3f", candidate.PBO.PBO, maxAllowedPBO)
	case !decision.TradeCountOK:
		decision.Reason = "trade count is below the configured minimum"
	case decision.SharpeDeltaVsBuyHold <= 0:
		decision.Reason = "Sharpe did not improve versus buy-hold"
	case decision.SortinoDeltaVsBuyHold <= 0:
		decision.Reason = "Sortino did not improve versus buy-hold"
	case decision.CalmarDeltaVsBuyHold <= 0:
		decision.Reason = "Calmar did not improve versus buy-hold"
	case decision.DrawdownDeltaVsBuyHold < -0.02:
		decision.Reason = "max drawdown regressed materially versus buy-hold"
	default:
		decision.Promote = true
		decision.Reason = "passed official promotion gates"
	}
	return decision
}
