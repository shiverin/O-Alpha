package portfolio

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/oalpha/internal/alpha/momentum"
	"github.com/oalpha/internal/alpha/ranker"
	"github.com/oalpha/internal/backtest"
)

const (
	defaultRankerArtifactRoot = "../reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts"
	defaultRankerVariant      = "stocks_h63_s15_top3_reb63_z10"
)

type StrategyRiskProfile string

const (
	StrategyRiskLow    StrategyRiskProfile = "low"
	StrategyRiskMedium StrategyRiskProfile = "medium"
	StrategyRiskHigh   StrategyRiskProfile = "high"
)

type StrategyDeploymentStatus string

const (
	StrategyStatusPromotedResearch    StrategyDeploymentStatus = "promoted_research_checkpoint"
	StrategyStatusConservativeVariant StrategyDeploymentStatus = "conservative_variant"
	StrategyStatusExperimentalVariant StrategyDeploymentStatus = "experimental_variant"
	StrategyStatusRejectedDiagnostic  StrategyDeploymentStatus = "rejected_diagnostic"
	StrategyStatusPaperOnly           StrategyDeploymentStatus = "paper_only"
)

type StrategyCatalogConfig struct {
	BenchmarkSymbol         string
	ModelArtifactRoot       string
	ModelVariant            string
	ModelYears              []int
	PointInTimeUniversePath string
}

type StrategySpec struct {
	Key                    string                   `json:"key"`
	DisplayName            string                   `json:"display_name"`
	Family                 string                   `json:"family"`
	RiskProfile            StrategyRiskProfile      `json:"risk_profile"`
	DeploymentStatus       StrategyDeploymentStatus `json:"deployment_status"`
	PromotedCheckpoint     bool                     `json:"promoted_checkpoint"`
	RequiresModelArtifacts bool                     `json:"requires_model_artifacts"`
	PaperOnly              bool                     `json:"paper_only"`
	BenchmarkSymbol        string                   `json:"benchmark_symbol"`
	Description            string                   `json:"description"`
	EvidencePaths          []string                 `json:"evidence_paths,omitempty"`
	Notes                  []string                 `json:"notes,omitempty"`
}

type strategyCatalogEntry struct {
	spec StrategySpec
	new  func([]string, StrategyCatalogConfig) backtest.PortfolioStrategy
}

func DefaultStrategyCatalogConfig() StrategyCatalogConfig {
	root := strings.TrimSpace(os.Getenv("OALPHA_DAILY_RANKER_ARTIFACT_ROOT"))
	if root == "" {
		root = defaultRankerArtifactRoot
	}
	return StrategyCatalogConfig{
		BenchmarkSymbol:         "VOO",
		ModelArtifactRoot:       filepath.Clean(root),
		ModelVariant:            defaultRankerVariant,
		ModelYears:              []int{2018, 2019, 2020, 2021, 2022, 2023, 2024, 2025, 2026},
		PointInTimeUniversePath: strings.TrimSpace(os.Getenv("OALPHA_DAILY_RANKER_PIT_UNIVERSE")),
	}
}

func AvailableStrategySpecs(symbols []string, cfg StrategyCatalogConfig) []StrategySpec {
	cfg = cfg.withDefaults(symbols)
	entries := catalogEntries(cfg)
	specs := make([]StrategySpec, 0, len(entries))
	for _, entry := range entries {
		specs = append(specs, entry.spec)
	}
	sort.Slice(specs, func(i, j int) bool {
		if specs[i].RiskProfile == specs[j].RiskProfile {
			return specs[i].Key < specs[j].Key
		}
		return riskOrder(specs[i].RiskProfile) < riskOrder(specs[j].RiskProfile)
	})
	return specs
}

func StrategySpecByKey(key string, symbols []string, cfg StrategyCatalogConfig) (StrategySpec, error) {
	cfg = cfg.withDefaults(symbols)
	normalizedKey := normalizeKey(key)
	for _, entry := range catalogEntries(cfg) {
		if entry.spec.Key == normalizedKey {
			return entry.spec, nil
		}
	}
	return StrategySpec{}, fmt.Errorf("unknown portfolio strategy %q", key)
}

func NewStrategyFromCatalog(key string, symbols []string, cfg StrategyCatalogConfig) (backtest.PortfolioStrategy, StrategySpec, error) {
	cfg = cfg.withDefaults(symbols)
	normalized := normalizeSymbols(symbols)
	if len(normalized) == 0 {
		return nil, StrategySpec{}, fmt.Errorf("at least one symbol is required")
	}
	normalizedKey := normalizeKey(key)
	for _, entry := range catalogEntries(cfg) {
		if entry.spec.Key != normalizedKey {
			continue
		}
		strategy := entry.new(normalized, cfg)
		if strategy == nil {
			return nil, StrategySpec{}, fmt.Errorf("strategy %s is not available for supplied symbols", normalizedKey)
		}
		return &labeledPortfolioStrategy{inner: strategy, spec: entry.spec}, entry.spec, nil
	}
	return nil, StrategySpec{}, fmt.Errorf("unknown portfolio strategy %q", key)
}

func (m *PortfolioAgentManager) StartCatalogPortfolioAgent(
	ctx context.Context,
	key string,
	strategyKey string,
	symbols []string,
	timeframe string,
	initialCash float64,
	cfg StrategyCatalogConfig,
	execution ExecutionRouter,
) (*PortfolioAgentWorker, StrategySpec, error) {
	strategy, spec, err := NewStrategyFromCatalog(strategyKey, symbols, cfg)
	if err != nil {
		return nil, StrategySpec{}, err
	}
	worker, err := m.StartPortfolioAgent(ctx, key, strategy, symbols, timeframe, initialCash, execution)
	if err != nil {
		return nil, StrategySpec{}, err
	}
	return worker, spec, nil
}

type labeledPortfolioStrategy struct {
	inner backtest.PortfolioStrategy
	spec  StrategySpec
}

func (s *labeledPortfolioStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	if s == nil || s.inner == nil {
		return nil, fmt.Errorf("labeled portfolio strategy requires an inner strategy")
	}
	outputs, err := s.inner.GeneratePortfolioSignals(ctx, panel)
	if err != nil {
		return nil, err
	}
	for i := range outputs {
		s.labelOutput(&outputs[i])
	}
	return outputs, nil
}

func (s *labeledPortfolioStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	if s == nil || s.inner == nil {
		return backtest.PortfolioOutput{}, fmt.Errorf("labeled portfolio strategy requires an inner strategy")
	}
	output, err := s.inner.EvaluatePortfolioLatest(ctx, panel)
	if err != nil {
		return backtest.PortfolioOutput{}, err
	}
	s.labelOutput(&output)
	return output, nil
}

func (s *labeledPortfolioStrategy) Universe() []string {
	if s == nil || s.inner == nil {
		return nil
	}
	return s.inner.Universe()
}

func (s *labeledPortfolioStrategy) Name() string {
	if s == nil {
		return ""
	}
	return s.spec.Key
}

func (s *labeledPortfolioStrategy) labelOutput(output *backtest.PortfolioOutput) {
	if output.EngineMetadata == nil {
		output.EngineMetadata = make(map[string]interface{})
	}
	output.EngineMetadata["agent_strategy_key"] = s.spec.Key
	output.EngineMetadata["agent_strategy_name"] = s.spec.DisplayName
	output.EngineMetadata["agent_strategy_family"] = s.spec.Family
	output.EngineMetadata["agent_strategy_risk_profile"] = string(s.spec.RiskProfile)
	output.EngineMetadata["agent_strategy_deployment_status"] = string(s.spec.DeploymentStatus)
	output.EngineMetadata["agent_strategy_promoted_checkpoint"] = s.spec.PromotedCheckpoint
	output.EngineMetadata["agent_strategy_paper_only"] = s.spec.PaperOnly
}

func catalogEntries(cfg StrategyCatalogConfig) []strategyCatalogEntry {
	benchmark := cfg.BenchmarkSymbol
	return []strategyCatalogEntry{
		{
			spec: StrategySpec{
				Key:                    "lgbm_ranker_h63_low",
				DisplayName:            "LGBM h63 active sleeve low risk",
				Family:                 "benchmark_lgbm_ranker_h63",
				RiskProfile:            StrategyRiskLow,
				DeploymentStatus:       StrategyStatusConservativeVariant,
				RequiresModelArtifacts: true,
				PaperOnly:              true,
				BenchmarkSymbol:        benchmark,
				Description:            "VOO core with a 5% learned-ranker active sleeve; stricter score threshold and single active name.",
				EvidencePaths:          h63EvidencePaths(),
				Notes: []string{
					"Conservative settings are derived from the promoted h63 checkpoint, but this exact sizing has not passed the official promotion gate.",
					"Research/paper only until PIT price coverage clears.",
				},
			},
			new: func(symbols []string, cfg StrategyCatalogConfig) backtest.PortfolioStrategy {
				return newDailyRanker(symbols, cfg, 0.05, 1, 0.05, 63, 1.50)
			},
		},
		{
			spec: StrategySpec{
				Key:              "ranker_proxy_h63_low",
				DisplayName:      "Deterministic h63 proxy low risk",
				Family:           "benchmark_ranker_proxy_h63",
				RiskProfile:      StrategyRiskLow,
				DeploymentStatus: StrategyStatusConservativeVariant,
				PaperOnly:        true,
				BenchmarkSymbol:  benchmark,
				Description:      "VOO core with an 8% deterministic h63 momentum/vol active sleeve.",
				EvidencePaths: []string{
					"reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_h63_longpanel_csv/",
				},
				Notes: []string{"Lower active sleeve than the official h63 proxy checkpoint; exact low-risk variant is not separately promoted."},
			},
			new: func(symbols []string, cfg StrategyCatalogConfig) backtest.PortfolioStrategy {
				return momentum.NewCompositeMomentumStrategy(symbols, rankerProxyConfig(cfg.BenchmarkSymbol, 63, 63, 0.08, 2, 0.04, 0.02, 0.35))
			},
		},
		{
			spec: StrategySpec{
				Key:              "lowvol_sleeve_low",
				DisplayName:      "Low-volatility sleeve low risk",
				Family:           "benchmark_lowvol",
				RiskProfile:      StrategyRiskLow,
				DeploymentStatus: StrategyStatusRejectedDiagnostic,
				PaperOnly:        true,
				BenchmarkSymbol:  benchmark,
				Description:      "VOO core with a small low-volatility stock sleeve.",
				EvidencePaths: []string{
					"reports/batches/2026-06-03_alpha_validation_yahoo100_longpanel_checkpoint/",
				},
				Notes: []string{"Low-vol sleeve is useful as a defensive comparison, but prior official runs rejected it as alpha."},
			},
			new: func(symbols []string, cfg StrategyCatalogConfig) backtest.PortfolioStrategy {
				return momentum.NewCompositeMomentumStrategy(symbols, lowVolConfig(cfg.BenchmarkSymbol, 42, 0.10, 5, "low_vol"))
			},
		},
		{
			spec: StrategySpec{
				Key:                    "lgbm_ranker_h63_medium",
				DisplayName:            "LGBM h63 active sleeve medium risk",
				Family:                 "benchmark_lgbm_ranker_h63",
				RiskProfile:            StrategyRiskMedium,
				DeploymentStatus:       StrategyStatusPromotedResearch,
				PromotedCheckpoint:     true,
				RequiresModelArtifacts: true,
				PaperOnly:              true,
				BenchmarkSymbol:        benchmark,
				Description:            "Current best checkpoint: VOO core with a 15% learned-ranker active sleeve, top 3 stocks, 63-bar rebalance.",
				EvidencePaths:          h63EvidencePaths(),
				Notes: []string{
					"Promoted in official Yahoo100 CSV windows 2015 through 2020 versus VOO.",
					"Still research/paper only because the current panel is survivorship-biased and PIT price coverage is not available.",
				},
			},
			new: func(symbols []string, cfg StrategyCatalogConfig) backtest.PortfolioStrategy {
				return newDailyRanker(symbols, cfg, 0.15, 3, 0.05, 63, 1.00)
			},
		},
		{
			spec: StrategySpec{
				Key:                "ranker_proxy_h63_medium",
				DisplayName:        "Deterministic h63 proxy medium risk",
				Family:             "benchmark_ranker_proxy_h63",
				RiskProfile:        StrategyRiskMedium,
				DeploymentStatus:   StrategyStatusPromotedResearch,
				PromotedCheckpoint: true,
				PaperOnly:          true,
				BenchmarkSymbol:    benchmark,
				Description:        "VOO core with the official deterministic h63 proxy active sleeve.",
				EvidencePaths: []string{
					"reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_h63_longpanel_csv/",
				},
				Notes: []string{"Primary 2015 window promoted, but shifted 2016 failed PBO; lower confidence than learned h63."},
			},
			new: func(symbols []string, cfg StrategyCatalogConfig) backtest.PortfolioStrategy {
				return momentum.NewCompositeMomentumStrategy(symbols, rankerProxyConfig(cfg.BenchmarkSymbol, 63, 63, 0.15, 3, 0.05, 0.00, 0.45))
			},
		},
		{
			spec: StrategySpec{
				Key:              "ranked_sleeve_medium",
				DisplayName:      "Risk-budgeted ranked sleeve medium risk",
				Family:           "benchmark_ranked_sleeve",
				RiskProfile:      StrategyRiskMedium,
				DeploymentStatus: StrategyStatusRejectedDiagnostic,
				PaperOnly:        true,
				BenchmarkSymbol:  benchmark,
				Description:      "VOO core with a broader risk-budgeted momentum sleeve.",
				EvidencePaths: []string{
					"reports/batches/2026-06-03_alpha_validation_yahoo100_ranked_sleeves_shifted_2017/",
					"reports/batches/2026-06-03_alpha_validation_yahoo100_ranked_sleeves_longpanel/",
				},
				Notes: []string{"Promoted in one shifted 2017 run, but failed the 2015 long-panel run; keep as diagnostic/fallback."},
			},
			new: func(symbols []string, cfg StrategyCatalogConfig) backtest.PortfolioStrategy {
				return momentum.NewCompositeMomentumStrategy(symbols, rankedSleeveConfig(cfg.BenchmarkSymbol, 21, 189, 0.30, 5, 0.08, 0.03))
			},
		},
		{
			spec: StrategySpec{
				Key:                    "lgbm_ranker_h63_high",
				DisplayName:            "LGBM h63 active sleeve high risk",
				Family:                 "benchmark_lgbm_ranker_h63",
				RiskProfile:            StrategyRiskHigh,
				DeploymentStatus:       StrategyStatusExperimentalVariant,
				RequiresModelArtifacts: true,
				PaperOnly:              true,
				BenchmarkSymbol:        benchmark,
				Description:            "VOO core with a 25% learned-ranker active sleeve, wider top-k, and lower score threshold.",
				EvidencePaths:          h63EvidencePaths(),
				Notes:                  []string{"Aggressive sizing variant is not promoted; use only for paper comparison and risk-cap testing."},
			},
			new: func(symbols []string, cfg StrategyCatalogConfig) backtest.PortfolioStrategy {
				return newDailyRanker(symbols, cfg, 0.25, 5, 0.07, 42, 0.50)
			},
		},
		{
			spec: StrategySpec{
				Key:              "benchmark_tsmom_high",
				DisplayName:      "Benchmark-funded TSMOM high risk",
				Family:           "benchmark_tsmom",
				RiskProfile:      StrategyRiskHigh,
				DeploymentStatus: StrategyStatusRejectedDiagnostic,
				PaperOnly:        true,
				BenchmarkSymbol:  benchmark,
				Description:      "VOO core with larger ETF and broad-universe time-series momentum sleeves.",
				EvidencePaths: []string{
					"reports/batches/2026-06-03_alpha_validation_yahoo100_benchmark_funded_suite_2015_csv/",
				},
				Notes: []string{"Economically plausible but rejected by PBO in official runs; useful as high-risk challenger only."},
			},
			new: func(symbols []string, cfg StrategyCatalogConfig) backtest.PortfolioStrategy {
				return momentum.NewCompositeMomentumStrategy(symbols, tsmomConfig(cfg.BenchmarkSymbol, 42, 126, 252, 0.18, 0.10))
			},
		},
		{
			spec: StrategySpec{
				Key:              "composite_momentum_high",
				DisplayName:      "Composite momentum high risk",
				Family:           "composite_momentum",
				RiskProfile:      StrategyRiskHigh,
				DeploymentStatus: StrategyStatusRejectedDiagnostic,
				PaperOnly:        true,
				BenchmarkSymbol:  benchmark,
				Description:      "Higher-active-weight composite momentum sleeve across ETFs and stocks.",
				EvidencePaths: []string{
					"reports/batches/2026-06-03_alpha_validation_yahoo100_longpanel_checkpoint/",
				},
				Notes: []string{"Raw return was attractive, but PBO failed; use only as high-risk experimental comparison."},
			},
			new: func(symbols []string, cfg StrategyCatalogConfig) backtest.PortfolioStrategy {
				return momentum.NewCompositeMomentumStrategy(symbols, compositeHighConfig(cfg.BenchmarkSymbol))
			},
		},
	}
}

func newDailyRanker(symbols []string, cfg StrategyCatalogConfig, sleeve float64, topK int, maxName float64, rebalance int, minScoreZ float64) backtest.PortfolioStrategy {
	return ranker.NewDailyRankerSleeveStrategy(symbols, ranker.DailyRankerSleeveConfig{
		BenchmarkSymbol:         cfg.BenchmarkSymbol,
		CandidateUniverse:       "stocks",
		PointInTimeUniversePath: cfg.PointInTimeUniversePath,
		ModelArtifactRoot:       cfg.ModelArtifactRoot,
		ModelVariant:            cfg.ModelVariant,
		ModelPathsByYear:        ranker.DailyRankerModelPaths(cfg.ModelArtifactRoot, cfg.ModelVariant, cfg.ModelYears...),
		RebalanceEveryBars:      rebalance,
		SleeveFraction:          sleeve,
		TopK:                    topK,
		MaxNameWeight:           maxName,
		TurnoverBand:            0.05,
		MinScoreZ:               minScoreZ,
	})
}

func rankerProxyConfig(benchmark string, rebalanceEvery int, lookback int, sleeve float64, topK int, maxName float64, minRelativeMomentum float64, maxVol20 float64) momentum.CompositeMomentumConfig {
	cfg := momentum.DefaultCompositeMomentumConfig()
	cfg.BenchmarkSymbol = benchmark
	cfg.RebalanceEveryBars = rebalanceEvery
	cfg.GlobalMaxNameWeight = maxName
	cfg.TurnoverBand = 0.05
	cfg.Legs = []momentum.CompositeMomentumLegConfig{
		{
			Name:                "ranker_proxy_stocks",
			CandidateUniverse:   "stocks",
			RankMode:            "vol_adjusted_momentum",
			WeightMode:          "risk_adjusted_edge",
			LookbackBars:        lookback,
			SleeveFraction:      sleeve,
			TopK:                topK,
			MaxNameWeight:       maxName,
			MinRelativeMomentum: minRelativeMomentum,
			MaxVol20:            maxVol20,
			EdgeExponent:        2,
			VolFloor:            0.10,
		},
	}
	return cfg
}

func rankedSleeveConfig(benchmark string, rebalanceEvery int, lookback int, sleeve float64, topK int, maxName float64, minRelativeMomentum float64) momentum.CompositeMomentumConfig {
	cfg := rankerProxyConfig(benchmark, rebalanceEvery, lookback, sleeve, topK, maxName, minRelativeMomentum, 0.45)
	if len(cfg.Legs) > 0 {
		cfg.Legs[0].Name = "risk_budgeted_stocks"
	}
	return cfg
}

func lowVolConfig(benchmark string, rebalanceEvery int, sleeve float64, topK int, rankMode string) momentum.CompositeMomentumConfig {
	cfg := momentum.DefaultCompositeMomentumConfig()
	cfg.BenchmarkSymbol = benchmark
	cfg.RebalanceEveryBars = rebalanceEvery
	cfg.GlobalMaxNameWeight = 0.10
	cfg.TurnoverBand = 0.05
	cfg.Legs = []momentum.CompositeMomentumLegConfig{
		{
			Name:                "stock_lowvol",
			CandidateUniverse:   "stocks",
			RankMode:            rankMode,
			LookbackBars:        63,
			SleeveFraction:      sleeve,
			TopK:                topK,
			MaxNameWeight:       sleeve / float64(topK),
			MinRelativeMomentum: -1,
			MaxVol20:            0.35,
		},
	}
	return cfg
}

func tsmomConfig(benchmark string, rebalanceEvery int, etfLookback int, broadLookback int, etfSleeve float64, broadSleeve float64) momentum.CompositeMomentumConfig {
	cfg := momentum.DefaultCompositeMomentumConfig()
	cfg.BenchmarkSymbol = benchmark
	cfg.RebalanceEveryBars = rebalanceEvery
	cfg.GlobalMaxNameWeight = 0.25
	cfg.TurnoverBand = 0.05
	cfg.Legs = []momentum.CompositeMomentumLegConfig{
		{
			Name:                "etf_tsmom",
			CandidateUniverse:   "etfs",
			LookbackBars:        etfLookback,
			SleeveFraction:      etfSleeve,
			TopK:                2,
			MaxNameWeight:       etfSleeve / 2,
			MinRelativeMomentum: 0.04,
			MaxVol20:            0.30,
		},
		{
			Name:                "broad_tsmom",
			CandidateUniverse:   "all",
			LookbackBars:        broadLookback,
			SleeveFraction:      broadSleeve,
			TopK:                5,
			MaxNameWeight:       broadSleeve / 5,
			MinRelativeMomentum: 0.08,
			MaxVol20:            0.45,
		},
	}
	return cfg
}

func compositeHighConfig(benchmark string) momentum.CompositeMomentumConfig {
	cfg := momentum.DefaultCompositeMomentumConfig()
	cfg.BenchmarkSymbol = benchmark
	cfg.RebalanceEveryBars = 21
	cfg.GlobalMaxNameWeight = 0.25
	cfg.TurnoverBand = 0.05
	cfg.Legs = []momentum.CompositeMomentumLegConfig{
		{
			Name:                "etf_21_high",
			CandidateUniverse:   "etfs",
			LookbackBars:        21,
			SleeveFraction:      0.24,
			TopK:                2,
			MaxNameWeight:       0.12,
			MinRelativeMomentum: 0.04,
			MaxVol20:            0.30,
		},
		{
			Name:                "stocks_126_high",
			CandidateUniverse:   "stocks",
			RankMode:            "vol_adjusted_momentum",
			WeightMode:          "risk_adjusted_edge",
			LookbackBars:        126,
			SleeveFraction:      0.16,
			TopK:                5,
			MaxNameWeight:       0.05,
			MinRelativeMomentum: 0.06,
			MaxVol20:            0.50,
			EdgeExponent:        2,
			VolFloor:            0.10,
		},
	}
	return cfg
}

func h63EvidencePaths() []string {
	return []string{
		"reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2015_csv/",
		"reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2016_csv/",
		"reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2017_csv/",
		"reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2018_csv/",
		"reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2019_csv/",
		"reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2020_csv/",
	}
}

func (c StrategyCatalogConfig) withDefaults(symbols []string) StrategyCatalogConfig {
	defaults := DefaultStrategyCatalogConfig()
	c.BenchmarkSymbol = strings.ToUpper(strings.TrimSpace(c.BenchmarkSymbol))
	if c.BenchmarkSymbol == "" {
		if len(symbols) > 0 {
			c.BenchmarkSymbol = strings.ToUpper(strings.TrimSpace(symbols[0]))
		}
		if c.BenchmarkSymbol == "" {
			c.BenchmarkSymbol = defaults.BenchmarkSymbol
		}
	}
	c.ModelArtifactRoot = strings.TrimSpace(c.ModelArtifactRoot)
	if c.ModelArtifactRoot == "" {
		c.ModelArtifactRoot = defaults.ModelArtifactRoot
	}
	c.ModelArtifactRoot = filepath.Clean(c.ModelArtifactRoot)
	c.ModelVariant = strings.TrimSpace(c.ModelVariant)
	if c.ModelVariant == "" {
		c.ModelVariant = defaults.ModelVariant
	}
	if len(c.ModelYears) == 0 {
		c.ModelYears = append([]int(nil), defaults.ModelYears...)
	}
	sort.Ints(c.ModelYears)
	c.PointInTimeUniversePath = strings.TrimSpace(c.PointInTimeUniversePath)
	return c
}

func normalizeKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeSymbols(symbols []string) []string {
	out := make([]string, 0, len(symbols))
	seen := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func riskOrder(profile StrategyRiskProfile) int {
	switch profile {
	case StrategyRiskLow:
		return 0
	case StrategyRiskMedium:
		return 1
	case StrategyRiskHigh:
		return 2
	default:
		return 3
	}
}
