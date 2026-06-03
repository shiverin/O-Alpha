package validation

import (
	"fmt"
	"math"
	"time"
)

type LabelWindow struct {
	EventTime time.Time `json:"event_time"`
	EndTime   time.Time `json:"end_time"`
}

type PurgedFold struct {
	Fold             int   `json:"fold"`
	TrainIndices     []int `json:"train_indices"`
	TestIndices      []int `json:"test_indices"`
	PurgedIndices    []int `json:"purged_indices"`
	EmbargoedIndices []int `json:"embargoed_indices"`
}

type PurgedCVReport struct {
	NumSamples int          `json:"num_samples"`
	NumFolds   int          `json:"num_folds"`
	EmbargoPct float64      `json:"embargo_pct"`
	Folds      []PurgedFold `json:"folds"`
}

func PurgedKFold(windows []LabelWindow, k int, embargoPct float64) ([]PurgedFold, error) {
	if len(windows) == 0 {
		return nil, fmt.Errorf("purged k-fold requires label windows")
	}
	if k < 2 {
		return nil, fmt.Errorf("k must be at least 2")
	}
	if k > len(windows) {
		return nil, fmt.Errorf("k=%d cannot exceed sample count %d", k, len(windows))
	}
	if embargoPct < 0 {
		embargoPct = 0
	}
	if embargoPct > 1 {
		embargoPct = 1
	}

	folds := make([]PurgedFold, 0, k)
	embargoCount := int(math.Ceil(float64(len(windows)) * embargoPct))
	for fold := 0; fold < k; fold++ {
		start := fold * len(windows) / k
		end := (fold + 1) * len(windows) / k
		testIndices := integerRange(start, end)
		testWindows := windows[start:end]
		embargoStart := end
		embargoEnd := minInt(len(windows), end+embargoCount)

		var trainIndices []int
		var purgedIndices []int
		var embargoedIndices []int
		for i, window := range windows {
			if i >= start && i < end {
				continue
			}
			if overlapsAny(window, testWindows) {
				purgedIndices = append(purgedIndices, i)
				continue
			}
			if i >= embargoStart && i < embargoEnd {
				embargoedIndices = append(embargoedIndices, i)
				continue
			}
			trainIndices = append(trainIndices, i)
		}

		folds = append(folds, PurgedFold{
			Fold:             fold,
			TrainIndices:     trainIndices,
			TestIndices:      testIndices,
			PurgedIndices:    purgedIndices,
			EmbargoedIndices: embargoedIndices,
		})
	}
	return folds, nil
}

func BuildPurgedCVReport(windows []LabelWindow, k int, embargoPct float64) (PurgedCVReport, error) {
	folds, err := PurgedKFold(windows, k, embargoPct)
	if err != nil {
		return PurgedCVReport{}, err
	}
	return PurgedCVReport{
		NumSamples: len(windows),
		NumFolds:   k,
		EmbargoPct: embargoPct,
		Folds:      folds,
	}, nil
}

func overlapsAny(window LabelWindow, candidates []LabelWindow) bool {
	for _, candidate := range candidates {
		if windowsOverlap(window, candidate) {
			return true
		}
	}
	return false
}

func windowsOverlap(a, b LabelWindow) bool {
	if a.EventTime.IsZero() || a.EndTime.IsZero() || b.EventTime.IsZero() || b.EndTime.IsZero() {
		return false
	}
	return a.EventTime.Before(b.EndTime) && b.EventTime.Before(a.EndTime)
}

func integerRange(start, end int) []int {
	out := make([]int, 0, maxInt(0, end-start))
	for i := start; i < end; i++ {
		out = append(out, i)
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
