package main

import (
	"testing"

	"github.com/oalpha/internal/db"
)

func TestNormalizedDatasetDefaults(t *testing.T) {
	got := normalizedDataset(db.BarQueryOptions{})
	if got["feed"] != "iex" || got["adjustment"] != "raw" || got["source"] != "alpaca" {
		t.Fatalf("unexpected defaults: %#v", got)
	}
}

func TestNormalizedDatasetUsesFlags(t *testing.T) {
	got := normalizedDataset(db.BarQueryOptions{
		Feed:       " Yahoo ",
		Adjustment: " Adj ",
		Source:     " Yahoo_Chart ",
	})
	if got["feed"] != "yahoo" || got["adjustment"] != "adj" || got["source"] != "yahoo_chart" {
		t.Fatalf("unexpected normalized dataset: %#v", got)
	}
}
