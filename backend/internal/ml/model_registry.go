package ml

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	ModelStatusCandidate = "candidate"
	ModelStatusPromoted  = "promoted"
	ModelStatusRejected  = "rejected"
	ModelTypeLightGBM    = "lightgbm_binary"
)

type ModelArtifact struct {
	ModelName               string                 `json:"model_name"`
	ModelType               string                 `json:"model_type"`
	StrategyScope           string                 `json:"strategy_scope"`
	ArtifactURI             string                 `json:"artifact_uri"`
	Manifest                ArtifactManifest       `json:"manifest,omitempty"`
	FeatureSpec             FeatureSpec            `json:"feature_spec"`
	LabelConfig             TripleBarrierConfig    `json:"label_config"`
	TrainingConfig          map[string]interface{} `json:"training_config"`
	Calibration             CalibrationModel       `json:"calibration,omitempty"`
	Thresholds              MLThresholds           `json:"thresholds,omitempty"`
	ThresholdSelection      map[string]interface{} `json:"threshold_selection,omitempty"`
	TrainStart              time.Time              `json:"train_start"`
	TrainEnd                time.Time              `json:"train_end"`
	ValidationStart         *time.Time             `json:"validation_start,omitempty"`
	ValidationEnd           *time.Time             `json:"validation_end,omitempty"`
	AUC                     float64                `json:"auc,omitempty"`
	LogLoss                 float64                `json:"logloss,omitempty"`
	SharpeNet               float64                `json:"sharpe_net,omitempty"`
	SortinoNet              float64                `json:"sortino_net,omitempty"`
	MaxDrawdownPct          float64                `json:"max_drawdown_pct,omitempty"`
	DSR                     float64                `json:"dsr,omitempty"`
	PBO                     float64                `json:"pbo,omitempty"`
	LeavesParityMaxAbsError float64                `json:"leaves_parity_max_abs_error,omitempty"`
	LeavesParityPassed      bool                   `json:"leaves_parity_passed"`
	Status                  string                 `json:"status"`
	CreatedAt               time.Time              `json:"created_at"`
}

type ArtifactManifest struct {
	ArtifactID        string                   `json:"artifact_id,omitempty"`
	GitSHA            string                   `json:"git_sha,omitempty"`
	ExportCommand     string                   `json:"export_command,omitempty"`
	TrainCommand      string                   `json:"train_command,omitempty"`
	BacktestCommand   string                   `json:"backtest_command,omitempty"`
	Symbols           []string                 `json:"symbols,omitempty"`
	ContextSymbols    []string                 `json:"context_symbols,omitempty"`
	Benchmark         string                   `json:"benchmark,omitempty"`
	FeatureSpecSHA256 string                   `json:"feature_spec_sha256,omitempty"`
	LabelConfigSHA256 string                   `json:"label_config_sha256,omitempty"`
	DataSnapshot      map[string]interface{}   `json:"data_snapshot,omitempty"`
	CostModel         map[string]interface{}   `json:"cost_model,omitempty"`
	Folds             []map[string]interface{} `json:"folds,omitempty"`
	ArtifactStatus    string                   `json:"artifact_status,omitempty"`
}

type ModelRegistry struct {
	RootDir string
}

func NewModelRegistry(rootDir string) *ModelRegistry {
	return &ModelRegistry{RootDir: rootDir}
}

func (r *ModelRegistry) LatestPromoted(modelName, strategyScope string) (*ModelArtifact, error) {
	artifacts, err := r.ListArtifacts()
	if err != nil {
		return nil, err
	}
	var matches []ModelArtifact
	for _, artifact := range artifacts {
		if artifact.Status != ModelStatusPromoted || !artifact.LeavesParityPassed {
			continue
		}
		if modelName != "" && artifact.ModelName != modelName {
			continue
		}
		if strategyScope != "" && artifact.StrategyScope != strategyScope {
			continue
		}
		matches = append(matches, artifact)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no promoted ML model artifact found for model=%q scope=%q", modelName, strategyScope)
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].CreatedAt.After(matches[j].CreatedAt)
	})
	return &matches[0], nil
}

func ResearchStatusAccepted(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case ModelStatusCandidate, ModelStatusPromoted:
		return true
	default:
		return false
	}
}

func (r *ModelRegistry) LoadLatestPromotedPredictor(modelName, strategyScope string) (*LeavesPredictor, *ModelArtifact, error) {
	artifact, err := r.LatestPromoted(modelName, strategyScope)
	if err != nil {
		return nil, nil, err
	}
	if artifact.ModelType != "" && artifact.ModelType != ModelTypeLightGBM {
		return nil, nil, fmt.Errorf("unsupported model type %q", artifact.ModelType)
	}
	predictor, err := NewLeavesPredictor(artifact.ModelPath(r.RootDir), artifact.FeatureSpec, artifact.Version())
	if err != nil {
		return nil, nil, err
	}
	return predictor, artifact, nil
}

func (r *ModelRegistry) ListArtifacts() ([]ModelArtifact, error) {
	root := strings.TrimSpace(r.RootDir)
	if root == "" {
		return nil, fmt.Errorf("model registry root is required")
	}
	var artifacts []ModelArtifact
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !isArtifactMetadataFile(path) {
			return nil
		}
		artifact, err := ReadModelArtifact(path)
		if err != nil {
			return fmt.Errorf("read model artifact %s: %w", path, err)
		}
		if artifact.CreatedAt.IsZero() {
			info, statErr := d.Info()
			if statErr == nil {
				artifact.CreatedAt = info.ModTime()
			}
		}
		artifacts = append(artifacts, artifact)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return artifacts, nil
}

func ReadModelArtifact(path string) (ModelArtifact, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ModelArtifact{}, err
	}
	var artifact ModelArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return ModelArtifact{}, err
	}
	return artifact.withDefaults(), nil
}

func WriteModelArtifact(path string, artifact ModelArtifact) error {
	artifact = artifact.withDefaults()
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func (a ModelArtifact) Version() string {
	if !a.CreatedAt.IsZero() {
		return fmt.Sprintf("%s@%s", a.ModelName, a.CreatedAt.UTC().Format(time.RFC3339))
	}
	if a.ArtifactURI != "" {
		return a.ArtifactURI
	}
	return a.ModelName
}

func (a ModelArtifact) ModelPath(rootDir string) string {
	uri := strings.TrimSpace(a.ArtifactURI)
	if uri == "" {
		return ""
	}
	if filepath.IsAbs(uri) {
		if isModelFile(uri) {
			return uri
		}
		return filepath.Join(uri, "model.txt")
	}
	path := filepath.Join(rootDir, uri)
	if isModelFile(path) {
		return path
	}
	return filepath.Join(path, "model.txt")
}

func (a ModelArtifact) withDefaults() ModelArtifact {
	if a.ModelType == "" {
		a.ModelType = ModelTypeLightGBM
	}
	if len(a.FeatureSpec.Features) == 0 {
		a.FeatureSpec = DefaultFeatureSpec()
	}
	if a.LabelConfig.HorizonBars == 0 {
		a.LabelConfig = a.LabelConfig.withDefaults()
	}
	if a.Thresholds.EnterLong <= 0 {
		a.Thresholds = DefaultMLThresholds()
	}
	if a.Status == "" {
		a.Status = ModelStatusCandidate
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	return a
}

func isArtifactMetadataFile(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return name == "metadata.json" || name == "model_artifact.json"
}

func isModelFile(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return name == "model.txt" || strings.HasSuffix(name, ".model") || strings.HasSuffix(name, ".txt")
}
