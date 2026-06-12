package api

import "testing"

func TestValidateOnboardingCompletionRequiresAcceptedBacktest(t *testing.T) {
	_, _, err := validateOnboardingCompletion(completeOnboardingRequest{
		RiskProfile: "moderate",
		StrategyKey: "lgbm_ranker_h63_medium",
	})
	if err == nil {
		t.Fatalf("expected missing acceptance error")
	}
}

func TestValidateOnboardingCompletionRejectsProfileMismatch(t *testing.T) {
	_, _, err := validateOnboardingCompletion(completeOnboardingRequest{
		RiskProfile:      "moderate",
		StrategyKey:      "composite_momentum_high",
		BacktestAccepted: true,
	})
	if err == nil {
		t.Fatalf("expected strategy/profile mismatch error")
	}
}

func TestValidateOnboardingCompletionAcceptsMatchingCatalogStrategy(t *testing.T) {
	riskProfile, strategyKey, err := validateOnboardingCompletion(completeOnboardingRequest{
		RiskProfile:      "Moderate",
		StrategyKey:      "LGBM_RANKER_H63_MEDIUM",
		BacktestAccepted: true,
	})
	if err != nil {
		t.Fatalf("validateOnboardingCompletion: %v", err)
	}
	if riskProfile != "moderate" {
		t.Fatalf("riskProfile=%s, want moderate", riskProfile)
	}
	if strategyKey != "lgbm_ranker_h63_medium" {
		t.Fatalf("strategyKey=%s, want lgbm_ranker_h63_medium", strategyKey)
	}
}
