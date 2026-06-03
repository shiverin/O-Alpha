package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oalpha/pkg/models"
)

func TestSignalValueParsesCommonEncodings(t *testing.T) {
	tests := map[string]models.Signal{
		"BUY":   models.SignalBuy,
		"LONG":  models.SignalBuy,
		"1":     models.SignalBuy,
		"SELL":  models.SignalSell,
		"SHORT": models.SignalSell,
		"-1":    models.SignalSell,
		"HOLD":  models.SignalHold,
		"":      models.SignalHold,
	}
	for input, want := range tests {
		if got := signalValue(input); got != want {
			t.Fatalf("signalValue(%q)=%v want %v", input, got, want)
		}
	}
}

func TestReadFeatureSpecYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "feature_spec.yaml")
	if err := os.WriteFile(path, []byte("version: test_v1\nfeatures:\n  - log_ret_1\nrequired_lookback: 64\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	spec, err := readFeatureSpec(path)
	if err != nil {
		t.Fatalf("readFeatureSpec: %v", err)
	}
	if spec.Version != "test_v1" || len(spec.Features) != 1 || spec.Features[0] != "log_ret_1" {
		t.Fatalf("unexpected spec: %+v", spec)
	}
}

func TestBarKeyNormalizesTimeToUTC(t *testing.T) {
	timestamp := time.Date(2024, 1, 2, 17, 0, 0, 0, time.FixedZone("SGT", 8*60*60))
	got := barKey(" voo ", timestamp)
	want := "VOO|2024-01-02T09:00:00Z"
	if got != want {
		t.Fatalf("barKey=%q want %q", got, want)
	}
}
