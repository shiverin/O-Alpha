package api

import (
	"testing"

	"github.com/oalpha/internal/agent"
)

func TestParseRegimeModeDefaultsToOverlay(t *testing.T) {
	mode, label, err := parseRegimeMode("", nil)
	if err != nil {
		t.Fatalf("parse regime mode failed: %v", err)
	}
	if mode != agent.RegimeModeOverlay || label != "overlay" {
		t.Fatalf("expected overlay default, got mode=%s label=%s", mode, label)
	}
}

func TestParseRegimeModeCanDisableOverlay(t *testing.T) {
	enabled := false
	mode, label, err := parseRegimeMode("", &enabled)
	if err != nil {
		t.Fatalf("parse regime mode failed: %v", err)
	}
	if mode != agent.RegimeModeNone || label != "none" {
		t.Fatalf("expected none when overlay disabled, got mode=%s label=%s", mode, label)
	}
}

func TestParseRegimeModeAcceptsExplicitNone(t *testing.T) {
	mode, label, err := parseRegimeMode("none", nil)
	if err != nil {
		t.Fatalf("parse regime mode failed: %v", err)
	}
	if mode != agent.RegimeModeNone || label != "none" {
		t.Fatalf("expected explicit none, got mode=%s label=%s", mode, label)
	}
}
