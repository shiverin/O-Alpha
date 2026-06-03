package ml

import (
	"path/filepath"
	"testing"
	"time"
)

func TestModelRegistryFindsLatestPromotedParityPassedArtifact(t *testing.T) {
	dir := t.TempDir()
	older := ModelArtifact{
		ModelName:          "meta_label",
		ModelType:          ModelTypeLightGBM,
		StrategyScope:      "ma_crossover",
		ArtifactURI:        "older",
		LeavesParityPassed: true,
		Status:             ModelStatusPromoted,
		CreatedAt:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	newer := older
	newer.ArtifactURI = "newer"
	newer.CreatedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rejected := newer
	rejected.ArtifactURI = "rejected"
	rejected.CreatedAt = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	rejected.LeavesParityPassed = false

	if err := WriteModelArtifact(filepath.Join(dir, "older", "metadata.json"), older); err != nil {
		t.Fatalf("write older: %v", err)
	}
	if err := WriteModelArtifact(filepath.Join(dir, "newer", "metadata.json"), newer); err != nil {
		t.Fatalf("write newer: %v", err)
	}
	if err := WriteModelArtifact(filepath.Join(dir, "rejected", "metadata.json"), rejected); err != nil {
		t.Fatalf("write rejected: %v", err)
	}

	registry := NewModelRegistry(dir)
	artifact, err := registry.LatestPromoted("meta_label", "ma_crossover")
	if err != nil {
		t.Fatalf("LatestPromoted: %v", err)
	}
	if artifact.ArtifactURI != "newer" {
		t.Fatalf("artifact=%s, want newer", artifact.ArtifactURI)
	}
}

func TestResearchStatusAcceptedIncludesCandidateAndPromoted(t *testing.T) {
	if !ResearchStatusAccepted(ModelStatusCandidate) {
		t.Fatalf("candidate should be research-accepted")
	}
	if !ResearchStatusAccepted(ModelStatusPromoted) {
		t.Fatalf("promoted should be research-accepted")
	}
	if ResearchStatusAccepted(ModelStatusRejected) {
		t.Fatalf("rejected should not be research-accepted")
	}
}
