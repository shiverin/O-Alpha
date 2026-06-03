package alphavalidation

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/oalpha/internal/backtest"
	btvalidation "github.com/oalpha/internal/backtest/validation"
)

type StrategyFactory struct {
	Name        string
	Family      string
	Benchmark   string
	AllowShorts bool
	New         func() backtest.PortfolioStrategy
}

type CostScenario struct {
	Name      string             `json:"name"`
	CostModel backtest.CostModel `json:"cost_model"`
}

type ValidationConfig struct {
	InitialCash      float64
	TrainBars        int
	TestBars         int
	StepBars         int
	MinOOSTrades     int
	NumTrials        int
	DataQualityPass  bool
	NoLookaheadPass  bool
	MaxGrossExposure float64
	MaxNetExposure   float64
	MaxSymbolWeight  float64
	CostScenarios    []CostScenario
	PromotionConfig  btvalidation.PromotionConfig
}

type AlphaValidationReport struct {
	GeneratedAt time.Time         `json:"generated_at"`
	Symbols     []string          `json:"symbols"`
	Timeframe   string            `json:"timeframe"`
	Start       time.Time         `json:"start"`
	End         time.Time         `json:"end"`
	BarCount    int               `json:"bar_count"`
	Config      ValidationConfig  `json:"config"`
	Benchmarks  []BenchmarkReport `json:"benchmarks"`
	Candidates  []CandidateReport `json:"candidates"`
	Notes       []string          `json:"notes"`
}

type BenchmarkReport struct {
	Name    string                    `json:"name"`
	Metrics backtest.PortfolioMetrics `json:"metrics"`
}

type CandidateReport struct {
	Name              string                         `json:"name"`
	Family            string                         `json:"family"`
	Benchmark         string                         `json:"benchmark"`
	Primary           ScenarioResult                 `json:"primary"`
	CostStress        []ScenarioResult               `json:"cost_stress"`
	WalkForward       []WindowResult                 `json:"walk_forward"`
	PBOEstimated      bool                           `json:"pbo_estimated"`
	PBODiagnostics    []PBODiagnostic                `json:"pbo_diagnostics,omitempty"`
	PromotionDecision btvalidation.PromotionDecision `json:"promotion_decision"`
	AuditMetadata     map[string]interface{}         `json:"audit_metadata,omitempty"`
	Diagnostics       []string                       `json:"diagnostics"`
}

type PBODiagnostic struct {
	Fold             int     `json:"fold"`
	TrainStart       int     `json:"train_start"`
	TrainEnd         int     `json:"train_end"`
	TestStart        int     `json:"test_start"`
	TestEnd          int     `json:"test_end"`
	Winner           string  `json:"winner"`
	WinnerTrainScore float64 `json:"winner_train_score"`
	WinnerTestScore  float64 `json:"winner_test_score"`
	WinnerTestRank   int     `json:"winner_test_rank"`
	TestWinner       string  `json:"test_winner"`
	TestWinnerScore  float64 `json:"test_winner_score"`
	VariantCount     int     `json:"variant_count"`
	Overfit          bool    `json:"overfit"`
}

type ScenarioResult struct {
	Scenario     string                    `json:"scenario"`
	Metrics      backtest.PortfolioMetrics `json:"metrics"`
	FinalEquity  float64                   `json:"final_equity"`
	NumTrades    int                       `json:"num_trades"`
	FeesPaid     float64                   `json:"fees_paid"`
	SlippageCost float64                   `json:"slippage_cost"`
	Error        string                    `json:"error,omitempty"`
}

type WindowResult struct {
	Fold       int                       `json:"fold"`
	TrainStart int                       `json:"train_start"`
	TrainEnd   int                       `json:"train_end"`
	TestStart  int                       `json:"test_start"`
	TestEnd    int                       `json:"test_end"`
	Train      backtest.PortfolioMetrics `json:"train"`
	Test       backtest.PortfolioMetrics `json:"test"`
	Error      string                    `json:"error,omitempty"`
}

func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		InitialCash:      100_000,
		TrainBars:        756,
		TestBars:         126,
		StepBars:         126,
		MinOOSTrades:     30,
		NumTrials:        1,
		DataQualityPass:  true,
		NoLookaheadPass:  true,
		MaxGrossExposure: 1,
		MaxNetExposure:   1,
		MaxSymbolWeight:  1,
		CostScenarios: []CostScenario{
			{Name: "normal", CostModel: backtest.DefaultCostModel()},
			{Name: "stress_2x", CostModel: scaledCostModel(backtest.DefaultCostModel(), 2)},
			{Name: "stress_3x", CostModel: scaledCostModel(backtest.DefaultCostModel(), 3)},
		},
		PromotionConfig: btvalidation.DefaultPromotionConfig(),
	}
}

func (r AlphaValidationReport) Markdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Alpha Validation Report\n\n")
	fmt.Fprintf(&b, "- Generated: `%s`\n", r.GeneratedAt.UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "- Symbols: `%s`\n", strings.Join(r.Symbols, ", "))
	fmt.Fprintf(&b, "- Timeframe: `%s`\n", r.Timeframe)
	fmt.Fprintf(&b, "- Period: `%s` to `%s`\n", r.Start.Format("2006-01-02"), r.End.Format("2006-01-02"))
	fmt.Fprintf(&b, "- Bars: `%d`\n\n", r.BarCount)

	if len(r.Notes) > 0 {
		fmt.Fprintf(&b, "## Notes\n\n")
		for _, note := range r.Notes {
			fmt.Fprintf(&b, "- %s\n", note)
		}
		fmt.Fprintf(&b, "\n")
	}

	fmt.Fprintf(&b, "## Benchmarks\n\n")
	fmt.Fprintf(&b, "| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |\n")
	fmt.Fprintf(&b, "|---|---:|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, benchmark := range r.Benchmarks {
		writeMetricRow(&b, benchmark.Name, benchmark.Metrics)
	}

	fmt.Fprintf(&b, "\n## Candidates\n\n")
	fmt.Fprintf(&b, "| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |\n")
	fmt.Fprintf(&b, "|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|\n")
	for _, candidate := range r.Candidates {
		metrics := candidate.Primary.Metrics
		reason := "pass"
		if len(candidate.PromotionDecision.Reasons) > 0 {
			reason = candidate.PromotionDecision.Reasons[0]
		}
		fmt.Fprintf(&b,
			"| %s | %s | %t | %.2f%% | %.2f%% | %.3f | %.3f | %.3f | %.2f%% | %.3f | %.3f | %d | %s |\n",
			candidate.Name,
			candidate.Benchmark,
			candidate.PromotionDecision.Promote,
			metrics.TotalReturn*100,
			metrics.AnnualReturn*100,
			metrics.Sharpe,
			metrics.Sortino,
			metrics.Calmar,
			metrics.MaxDrawdown*100,
			metrics.DSR,
			metrics.PBO,
			metrics.NumTrades,
			escapePipes(reason),
		)
	}

	for _, candidate := range r.Candidates {
		fmt.Fprintf(&b, "\n## %s\n\n", candidate.Name)
		fmt.Fprintf(&b, "- Family: `%s`\n", candidate.Family)
		fmt.Fprintf(&b, "- Benchmark: `%s`\n", candidate.Benchmark)
		fmt.Fprintf(&b, "- PBO estimated: `%t`\n", candidate.PBOEstimated)
		fmt.Fprintf(&b, "- Promotion: `%t`\n", candidate.PromotionDecision.Promote)
		if len(candidate.PromotionDecision.Reasons) > 0 {
			fmt.Fprintf(&b, "- Rejection reasons:\n")
			for _, reason := range candidate.PromotionDecision.Reasons {
				fmt.Fprintf(&b, "  - %s\n", reason)
			}
		}
		if len(candidate.Diagnostics) > 0 {
			fmt.Fprintf(&b, "- Diagnostics:\n")
			for _, diagnostic := range candidate.Diagnostics {
				fmt.Fprintf(&b, "  - %s\n", diagnostic)
			}
		}
		if len(candidate.AuditMetadata) > 0 {
			fmt.Fprintf(&b, "\n### Metadata Audit\n\n")
			fmt.Fprintf(&b, "| Key | Value |\n")
			fmt.Fprintf(&b, "|---|---|\n")
			keys := make([]string, 0, len(candidate.AuditMetadata))
			for key := range candidate.AuditMetadata {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				fmt.Fprintf(&b, "| %s | `%s` |\n", key, escapePipes(formatMetadataValue(candidate.AuditMetadata[key])))
			}
		}

		fmt.Fprintf(&b, "\n### Cost Stress\n\n")
		fmt.Fprintf(&b, "| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |\n")
		fmt.Fprintf(&b, "|---|---:|---:|---:|---:|---:|---:|---:|---:|---|\n")
		for _, scenario := range candidate.CostStress {
			errText := scenario.Error
			if errText == "" {
				errText = "-"
			}
			fmt.Fprintf(&b, "| %s | %.2f%% | %.2f%% | %.3f | %.3f | %.3f | %.2f%% | %d | %.3f | %s |\n",
				scenario.Scenario,
				scenario.Metrics.TotalReturn*100,
				scenario.Metrics.AnnualReturn*100,
				scenario.Metrics.Sharpe,
				scenario.Metrics.Sortino,
				scenario.Metrics.Calmar,
				scenario.Metrics.MaxDrawdown*100,
				scenario.Metrics.NumTrades,
				scenario.Metrics.Turnover,
				escapePipes(errText),
			)
		}

		if len(candidate.WalkForward) > 0 {
			fmt.Fprintf(&b, "\n### Walk Forward\n\n")
			fmt.Fprintf(&b, "| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |\n")
			fmt.Fprintf(&b, "|---:|---:|---:|---:|---:|---:|---:|---|\n")
			for _, window := range candidate.WalkForward {
				errText := window.Error
				if errText == "" {
					errText = "-"
				}
				fmt.Fprintf(&b, "| %d | %d-%d | %d-%d | %.3f | %.3f | %.3f | %.2f%% | %s |\n",
					window.Fold,
					window.TrainStart,
					window.TrainEnd,
					window.TestStart,
					window.TestEnd,
					window.Train.Sharpe,
					window.Test.Sharpe,
					window.Test.Calmar,
					window.Test.MaxDrawdown*100,
					escapePipes(errText),
				)
			}
		}

		if len(candidate.PBODiagnostics) > 0 {
			fmt.Fprintf(&b, "\n### PBO Diagnostics\n\n")
			fmt.Fprintf(&b, "| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |\n")
			fmt.Fprintf(&b, "|---:|---|---:|---:|---:|---|---:|---:|---:|\n")
			for _, row := range candidate.PBODiagnostics {
				fmt.Fprintf(&b, "| %d | %s | %.3f | %.3f | %d | %s | %.3f | %d | %t |\n",
					row.Fold,
					row.Winner,
					row.WinnerTrainScore,
					row.WinnerTestScore,
					row.WinnerTestRank,
					row.TestWinner,
					row.TestWinnerScore,
					row.VariantCount,
					row.Overfit,
				)
			}
		}
	}
	return b.String()
}

func writeMetricRow(b *strings.Builder, name string, metrics backtest.PortfolioMetrics) {
	fmt.Fprintf(b,
		"| %s | %.2f%% | %.2f%% | %.3f | %.3f | %.3f | %.2f%% | %d | %.3f |\n",
		name,
		metrics.TotalReturn*100,
		metrics.AnnualReturn*100,
		metrics.Sharpe,
		metrics.Sortino,
		metrics.Calmar,
		metrics.MaxDrawdown*100,
		metrics.NumTrades,
		metrics.Turnover,
	)
}

func escapePipes(value string) string {
	return strings.ReplaceAll(value, "|", "\\|")
}

func formatMetadataValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case bool:
		return fmt.Sprintf("%t", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.10g", v)
	case []string:
		return strings.Join(v, ",")
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, formatMetadataValue(item))
		}
		return strings.Join(parts, ",")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func scaledCostModel(model backtest.CostModel, scalar float64) backtest.CostModel {
	if scalar <= 0 || math.IsNaN(scalar) || math.IsInf(scalar, 0) {
		scalar = 1
	}
	model.DefaultSpreadBps *= scalar
	model.SlippageBps *= scalar
	model.CommissionPerShare *= scalar
	model.MinCommission *= scalar
	model.BorrowFeeBpsAnnual *= scalar
	model.SECFeesBpsSell *= scalar
	if model.SpreadBpsBySymbol != nil {
		overrides := make(map[string]float64, len(model.SpreadBpsBySymbol))
		for symbol, value := range model.SpreadBpsBySymbol {
			overrides[symbol] = value * scalar
		}
		model.SpreadBpsBySymbol = overrides
	}
	return model
}
