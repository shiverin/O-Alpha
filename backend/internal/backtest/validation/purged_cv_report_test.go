package validation

import (
	"slices"
	"testing"
	"time"
)

func TestPurgedKFoldRemovesOverlappingAndEmbargoedSamples(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	windows := make([]LabelWindow, 6)
	for i := range windows {
		windows[i] = LabelWindow{
			EventTime: start.AddDate(0, 0, i),
			EndTime:   start.AddDate(0, 0, i+1),
		}
	}
	windows[3] = LabelWindow{
		EventTime: start.AddDate(0, 0, 1),
		EndTime:   start.AddDate(0, 0, 4),
	}

	folds, err := PurgedKFold(windows, 3, 0.17)
	if err != nil {
		t.Fatalf("PurgedKFold: %v", err)
	}
	first := folds[0]
	if !slices.Contains(first.PurgedIndices, 3) {
		t.Fatalf("purged=%v, want overlapping sample 3 purged", first.PurgedIndices)
	}
	if !slices.Contains(first.EmbargoedIndices, 2) {
		t.Fatalf("embargoed=%v, want sample 2 embargoed", first.EmbargoedIndices)
	}
	if slices.Contains(first.TrainIndices, 0) || slices.Contains(first.TrainIndices, 1) {
		t.Fatalf("train includes test samples: %v", first.TrainIndices)
	}
}
