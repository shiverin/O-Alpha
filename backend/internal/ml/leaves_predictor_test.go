package ml

import (
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestLeavesPredictorProbabilityAndRawOutputs(t *testing.T) {
	modelPath := writeLightGBMOneLeafModel(t, "binary sigmoid:1", 0.123)
	predictor, err := NewLeavesPredictor(modelPath, FeatureSpec{
		Version:  "test",
		Features: []string{"f0"},
	}, "test-model")
	if err != nil {
		t.Fatalf("NewLeavesPredictor: %v", err)
	}

	probability, err := predictor.PredictProba([]float64{0})
	if err != nil {
		t.Fatalf("PredictProba: %v", err)
	}
	expectedProbability := 1 / (1 + math.Exp(-0.123))
	if math.Abs(probability-expectedProbability) > 1e-12 {
		t.Fatalf("probability=%0.15f, want %0.15f", probability, expectedProbability)
	}

	raw, err := predictor.PredictRaw([]float64{0})
	if err != nil {
		t.Fatalf("PredictRaw: %v", err)
	}
	if math.Abs(raw-0.123) > 1e-12 {
		t.Fatalf("raw=%0.15f, want 0.123", raw)
	}
}

func TestRawLeavesPredictorLoadsLambdaRankObjective(t *testing.T) {
	modelPath := writeLightGBMOneLeafModel(t, "lambdarank", 0.456)
	if _, err := NewLeavesPredictor(modelPath, FeatureSpec{
		Version:  "test",
		Features: []string{"f0"},
	}, "test-model"); err == nil {
		t.Fatalf("NewLeavesPredictor unexpectedly loaded lambdarank with transformations enabled")
	}

	predictor, err := NewRawLeavesPredictor(modelPath, FeatureSpec{
		Version:  "test",
		Features: []string{"f0"},
	}, "test-model")
	if err != nil {
		t.Fatalf("NewRawLeavesPredictor: %v", err)
	}

	raw, err := predictor.PredictRaw([]float64{0})
	if err != nil {
		t.Fatalf("PredictRaw: %v", err)
	}
	if math.Abs(raw-0.456) > 1e-12 {
		t.Fatalf("raw=%0.15f, want 0.456", raw)
	}

	if _, err := predictor.PredictProba([]float64{0}); err == nil {
		t.Fatalf("PredictProba unexpectedly succeeded for raw-score predictor")
	}
}

func writeLightGBMOneLeafModel(t *testing.T, objective string, leafValue float64) string {
	t.Helper()
	payload := strings.Join([]string{
		"tree",
		"version=v2",
		"num_class=1",
		"num_tree_per_iteration=1",
		"label_index=0",
		"max_feature_idx=0",
		"objective=" + objective,
		"feature_names=f0",
		"feature_infos=[0:1]",
		"tree_sizes=128",
		"",
		"Tree=0",
		"num_leaves=1",
		"num_cat=0",
		"split_feature=",
		"split_gain=",
		"threshold=",
		"decision_type=",
		"left_child=",
		"right_child=",
		"leaf_value=" + strconvFloat(leafValue),
		"leaf_count=",
		"internal_value=",
		"internal_count=",
		"shrinkage=1",
		"",
		"end of trees",
		"",
	}, "\n")
	path := filepath.Join(t.TempDir(), "model.txt")
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("write model: %v", err)
	}
	return path
}

func strconvFloat(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}
