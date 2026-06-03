package cointegration

import (
	"context"
	"testing"
	"time"
)

func TestPairCandidateRejectsCostInfeasibleAndNearIdenticalETF(t *testing.T) {
	cfg := DefaultPairSelectionConfig()
	costly := approvedCandidate()
	costly.AvgSpreadBps = 10
	costly.EstimatedRoundTripCostBps = 5
	if ok, reason := costly.Passes(cfg); ok || reason != "spread_move_not_cost_feasible" {
		t.Fatalf("costly ok=%v reason=%s, want spread_move_not_cost_feasible", ok, reason)
	}

	nearIdentical := approvedCandidate()
	nearIdentical.NearIdenticalETF = true
	if ok, reason := nearIdentical.Passes(cfg); ok || reason != "near_identical_etf_rejected" {
		t.Fatalf("near-identical ok=%v reason=%s, want near_identical_etf_rejected", ok, reason)
	}
}

func TestStaticPairCandidateRepositoryReturnsLatestApproved(t *testing.T) {
	old := approvedCandidate()
	old.FormationEnd = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := approvedCandidate()
	newer.FormationEnd = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := StaticPairCandidateRepository{Candidates: []PairCandidate{old, newer}}

	candidate, err := repo.LatestApprovedPair(context.Background(), "Y", "X", "1Day")
	if err != nil {
		t.Fatalf("LatestApprovedPair: %v", err)
	}
	if !candidate.FormationEnd.Equal(newer.FormationEnd) {
		t.Fatalf("formation_end=%s, want latest %s", candidate.FormationEnd, newer.FormationEnd)
	}
}

func approvedCandidate() PairCandidate {
	return PairCandidate{
		SymbolY:                   "Y",
		SymbolX:                   "X",
		Timeframe:                 "1Day",
		FormationEnd:              time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Correlation:               0.9,
		EngleGrangerPValue:        0.005,
		HalfLifeBars:              10,
		AvgSpreadBps:              30,
		EstimatedRoundTripCostBps: 5,
		Approved:                  true,
		Status:                    "approved",
	}
}
