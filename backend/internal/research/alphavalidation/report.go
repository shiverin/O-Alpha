package alphavalidation

import (
	"fmt"
	"strings"
	"time"

	"github.com/oalpha/pkg/models"
)

type ValidationWindow struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type CostScenarioResult struct {
	Name             string  `json:"name"`
	TransactionCostBPS float64 `json:"transaction_cost_bps"`
	FinalEquity      float64 `json:"final_equity"`
	TotalReturn      float64 `json:"total_return"`
	AnnualizedReturn float64 `json:"annualized_return"`
	Sharpe           float64 `json:"sharpe"`
	Sortino          float64 `json:"sortino"`
	Calmar           float64 `json:"calmar"`
	MaxDrawdown      float64 `json:"max_drawdown"`
	NumTrades        int     `json:"num_trades"`
	Turnover         float64 `json:"turnover"`
	FeesPaid         float64 `json:"fees_paid"`
}

type FoldResult struct {
	Fold     int                    `json:"fold"`
	Train    ValidationWindow       `json:"train"`
	Test     ValidationWindow       `json:"test"`
	Score    float64                `json:"score"`
	Result   *models.BacktestResult `json:"result"`
	BuyHold  *models.BacktestResult `json:"buy_hold"`
}

type PBODiagnostics struct {
	Estimated          bool      `json:"estimated"`
	Method             string    `json:"method"`
	Objective          string    `json:"objective"`
	VariantCount       int       `json:"variant_count"`
	FoldCount          int       `json:"fold_count"`
	SplitCount         int       `json:"split_count"`
	PBO                float64   `json:"pbo"`
	MedianLambda       float64   `json:"median_lambda"`
	TrainWinnerCounts  map[string]int `json:"train_winner_counts,omitempty"`
	FailureReason      string    `json:"failure_reason,omitempty"`
}

type PromotionDecision struct {
	Promote                 bool    `json:"promote"`
	Reason                  string  `json:"reason"`
	PBOEstimated            bool    `json:"pbo_estimated"`
	PBO                     float64 `json:"pbo"`
	SharpeDeltaVsBuyHold    float64 `json:"sharpe_delta_vs_buy_hold"`
	SortinoDeltaVsBuyHold   float64 `json:"sortino_delta_vs_buy_hold"`
	CalmarDeltaVsBuyHold    float64 `json:"calmar_delta_vs_buy_hold"`
	DrawdownDeltaVsBuyHold  float64 `json:"drawdown_delta_vs_buy_hold"`
	TradeCountOK            bool    `json:"trade_count_ok"`
}

type CandidateReport struct {
	Strategy            string                        `json:"strategy"`
	Family              string                        `json:"family"`
	Variant             string                        `json:"variant"`
	Description         string                        `json:"description"`
	BenchmarkSymbol     string                        `json:"benchmark_symbol"`
	ActiveSymbols       []string                      `json:"active_symbols"`
	LookbackBars        []int                         `json:"lookback_bars"`
	LookbackWeights     []float64                     `json:"lookback_weights"`
	MaxPositions        int                           `json:"max_positions"`
	RebalanceBars       int                           `json:"rebalance_bars"`
	TurnoverBufferRanks int                           `json:"turnover_buffer_ranks"`
	ActiveSleevePct     float64                       `json:"active_sleeve_pct"`
	TransactionCostBPS  float64                       `json:"transaction_cost_bps"`
	Result              *models.BacktestResult        `json:"result"`
	CostScenarios       []CostScenarioResult          `json:"cost_scenarios,omitempty"`
	FoldResults         []FoldResult                  `json:"fold_results,omitempty"`
	PBO                 *PBODiagnostics               `json:"pbo_diagnostics,omitempty"`
	Promotion           PromotionDecision             `json:"promotion"`
}

type ValidationReport struct {
	GeneratedAt      time.Time         `json:"generated_at"`
	StrategyFamily   string            `json:"strategy_family"`
	RequestedStrategy string           `json:"requested_strategy"`
	BenchmarkSymbol  string            `json:"benchmark_symbol"`
	Symbols          []string          `json:"symbols"`
	Timeframe        string            `json:"timeframe"`
	Window           ValidationWindow  `json:"window"`
	TrainBars        int               `json:"train_bars"`
	TestBars         int               `json:"test_bars"`
	StepBars         int               `json:"step_bars"`
	MinTrades        int               `json:"min_trades"`
	InitialCash      float64           `json:"initial_cash"`
	BuyHold          *models.BacktestResult `json:"buy_hold"`
	Candidates       []CandidateReport `json:"candidates"`
}

func RenderMarkdown(report ValidationReport, jsonPath string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Alpha Validation Report\n\n")
	fmt.Fprintf(&b, "- Generated: `%s`\n", report.GeneratedAt.UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "- Strategy family: `%s`\n", report.StrategyFamily)
	fmt.Fprintf(&b, "- Requested strategy: `%s`\n", report.RequestedStrategy)
	fmt.Fprintf(&b, "- Benchmark: `%s`\n", report.BenchmarkSymbol)
	fmt.Fprintf(&b, "- Timeframe: `%s`\n", report.Timeframe)
	fmt.Fprintf(&b, "- Window: `%s` to `%s`\n", report.Window.From.Format("2006-01-02"), report.Window.To.Format("2006-01-02"))
	fmt.Fprintf(&b, "- Symbols: `%s`\n", strings.Join(report.Symbols, ","))
	fmt.Fprintf(&b, "- JSON report: `%s`\n\n", jsonPath)

	if report.BuyHold != nil {
		fmt.Fprintf(&b, "## Buy-and-hold benchmark\n\n")
		fmt.Fprintf(&b, "- Return: `%.2f%%`\n", report.BuyHold.TotalReturn*100)
		fmt.Fprintf(&b, "- Sharpe: `%.3f`\n", report.BuyHold.Sharpe)
		fmt.Fprintf(&b, "- Sortino: `%.3f`\n", report.BuyHold.Sortino)
		fmt.Fprintf(&b, "- Calmar: `%.3f`\n", report.BuyHold.Calmar)
		fmt.Fprintf(&b, "- Max drawdown: `%.2f%%`\n", report.BuyHold.MaxDrawdown*100)
		fmt.Fprintf(&b, "- Trades: `%d`\n\n", report.BuyHold.NumTrades)
	}

	fmt.Fprintf(&b, "## Candidates\n\n")
	for _, candidate := range report.Candidates {
		fmt.Fprintf(&b, "### `%s`\n\n", candidate.Strategy)
		fmt.Fprintf(&b, "- Description: %s\n", candidate.Description)
		fmt.Fprintf(&b, "- Active sleeve: `%.2f%%`\n", candidate.ActiveSleevePct*100)
		fmt.Fprintf(&b, "- Lookbacks: `%v`\n", candidate.LookbackBars)
		fmt.Fprintf(&b, "- Rebalance bars: `%d`\n", candidate.RebalanceBars)
		fmt.Fprintf(&b, "- Max positions: `%d`\n", candidate.MaxPositions)
		if candidate.Result != nil {
			fmt.Fprintf(&b, "- Result: return `%.2f%%` | Sharpe `%.3f` | Sortino `%.3f` | Calmar `%.3f` | max drawdown `%.2f%%` | trades `%d` | promote? `%t` | first gate reason `%s`\n",
				candidate.Result.TotalReturn*100,
				candidate.Result.Sharpe,
				candidate.Result.Sortino,
				candidate.Result.Calmar,
				candidate.Result.MaxDrawdown*100,
				candidate.Result.NumTrades,
				candidate.Promotion.Promote,
				candidate.Promotion.Reason,
			)
		}
		if candidate.PBO != nil {
			if candidate.PBO.Estimated {
				fmt.Fprintf(&b, "- PBO: `%.3f` from `%d` variants over `%d` folds\n", candidate.PBO.PBO, candidate.PBO.VariantCount, candidate.PBO.FoldCount)
			} else {
				fmt.Fprintf(&b, "- PBO: not estimated (`%s`)\n", candidate.PBO.FailureReason)
			}
		}
		if len(candidate.CostScenarios) > 0 {
			fmt.Fprintf(&b, "- Cost stress:\n")
			for _, stress := range candidate.CostScenarios {
				fmt.Fprintf(&b, "  - `%s`: return `%.2f%%` | Sharpe `%.3f` | max drawdown `%.2f%%`\n", stress.Name, stress.TotalReturn*100, stress.Sharpe, stress.MaxDrawdown*100)
			}
		}
		fmt.Fprintln(&b)
	}
	return b.String()
}
