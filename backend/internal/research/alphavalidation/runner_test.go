package alphavalidation

import (
	"context"
<<<<<<< HEAD
	"strings"
	"testing"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

func TestRunValidationFailsClosedWhenPBOUnavailable(t *testing.T) {
	panel := validationTestPanel(80)
	cfg := DefaultValidationConfig()
	cfg.TrainBars = 20
	cfg.TestBars = 10
	cfg.StepBars = 10
	cfg.MinOOSTrades = 1

	report, err := RunValidation(context.Background(), panel, []StrategyFactory{flatFactory()}, []StrategyFactory{toggleFactory("toggle_once", "custom", false)}, cfg)
	if err != nil {
		t.Fatalf("RunValidation: %v", err)
	}
	if len(report.Candidates) != 1 {
		t.Fatalf("candidates=%d, want 1", len(report.Candidates))
	}
	candidate := report.Candidates[0]
	if candidate.PBOEstimated {
		t.Fatalf("PBO should not be estimated for a single custom variant")
	}
	if candidate.PromotionDecision.Promote {
		t.Fatalf("candidate should fail closed without PBO")
	}
	if !containsReason(candidate.PromotionDecision.Reasons, "PBO was not estimated") {
		t.Fatalf("reasons=%v, want PBO failure", candidate.PromotionDecision.Reasons)
	}
}

func TestRunValidationCostStressIsVisible(t *testing.T) {
	panel := validationTestPanel(80)
	cfg := DefaultValidationConfig()
	cfg.TrainBars = 20
	cfg.TestBars = 10
	cfg.StepBars = 10
	cfg.MinOOSTrades = 1

	report, err := RunValidation(context.Background(), panel, []StrategyFactory{flatFactory()}, []StrategyFactory{toggleFactory("toggle_daily", "custom", true)}, cfg)
	if err != nil {
		t.Fatalf("RunValidation: %v", err)
	}
	stress := report.Candidates[0].CostStress
	if len(stress) < 3 {
		t.Fatalf("cost stress scenarios=%d, want at least 3", len(stress))
	}
	if stress[2].FinalEquity >= stress[0].FinalEquity {
		t.Fatalf("3x stress final equity=%f should be below normal=%f", stress[2].FinalEquity, stress[0].FinalEquity)
	}
}

func TestAlphaValidationMarkdownHasCandidateTable(t *testing.T) {
	panel := validationTestPanel(40)
	cfg := DefaultValidationConfig()
	cfg.TrainBars = 10
	cfg.TestBars = 5
	cfg.StepBars = 5
	report, err := RunValidation(context.Background(), panel, []StrategyFactory{flatFactory()}, []StrategyFactory{toggleFactory("toggle", "custom", false)}, cfg)
	if err != nil {
		t.Fatalf("RunValidation: %v", err)
	}
	markdown := report.Markdown()
	if !strings.Contains(markdown, "## Candidates") || !strings.Contains(markdown, "toggle") {
		t.Fatalf("markdown missing candidate table:\n%s", markdown)
	}
}

func TestRunValidationIncludesCandidateAuditMetadata(t *testing.T) {
	panel := validationTestPanel(40)
	cfg := DefaultValidationConfig()
	cfg.TrainBars = 10
	cfg.TestBars = 5
	cfg.StepBars = 5

	report, err := RunValidation(context.Background(), panel, []StrategyFactory{flatFactory()}, []StrategyFactory{metadataFactory()}, cfg)
	if err != nil {
		t.Fatalf("RunValidation: %v", err)
	}
	metadata := report.Candidates[0].AuditMetadata
	if metadata["ranker_model_sha256"] != "abc123" {
		t.Fatalf("ranker_model_sha256=%v, want abc123; metadata=%v", metadata["ranker_model_sha256"], metadata)
	}
	if _, ok := metadata["selection_rows"]; ok {
		t.Fatalf("selection_rows should be omitted from compact audit metadata: %v", metadata)
	}
	if markdown := report.Markdown(); !strings.Contains(markdown, "### Metadata Audit") || !strings.Contains(markdown, "ranker_model_sha256") {
		t.Fatalf("markdown missing metadata audit:\n%s", markdown)
	}
}

func TestWalkForwardTestUsesTrainWindowAsWarmup(t *testing.T) {
	panel := validationTestPanel(50)
	cfg := DefaultValidationConfig()
	cfg.TrainBars = 20
	cfg.TestBars = 10
	cfg.StepBars = 10
	cfg.MinOOSTrades = 1

	report, err := RunValidation(context.Background(), panel, []StrategyFactory{flatFactory()}, []StrategyFactory{warmupToggleFactory(15)}, cfg)
	if err != nil {
		t.Fatalf("RunValidation: %v", err)
	}
	if len(report.Candidates) != 1 || len(report.Candidates[0].WalkForward) == 0 {
		t.Fatalf("missing candidate walk-forward results: %#v", report.Candidates)
	}
	testMetrics := report.Candidates[0].WalkForward[0].Test
	if testMetrics.NumTrades == 0 {
		t.Fatalf("expected OOS trades after warmup context, got metrics=%+v", testMetrics)
	}
}

func TestRunValidationIncludesPBODiagnostics(t *testing.T) {
	panel := validationTestPanel(80)
	cfg := DefaultValidationConfig()
	cfg.TrainBars = 20
	cfg.TestBars = 10
	cfg.StepBars = 10
	cfg.MinOOSTrades = 1

	report, err := RunValidation(context.Background(), panel, []StrategyFactory{BenchmarkFactories(panel.Symbols)[0]}, []StrategyFactory{maFactory("VOO", 2, 5, "ma_test")}, cfg)
	if err != nil {
		t.Fatalf("RunValidation: %v", err)
	}
	if len(report.Candidates) != 1 {
		t.Fatalf("candidates=%d, want 1", len(report.Candidates))
	}
	if len(report.Candidates[0].PBODiagnostics) == 0 {
		t.Fatalf("expected PBO diagnostics for MA family")
	}
	for _, row := range report.Candidates[0].PBODiagnostics {
		if row.Winner == "" || row.VariantCount < 2 || row.WinnerTestRank <= 0 {
			t.Fatalf("invalid PBO diagnostic row: %+v", row)
		}
	}
}

func flatFactory() StrategyFactory {
	return StrategyFactory{
		Name:      "flat_cash",
		Family:    "benchmark",
		Benchmark: "",
		New: func() backtest.PortfolioStrategy {
			return &flatStrategy{symbols: []string{"VOO"}}
		},
	}
}

func toggleFactory(name, family string, daily bool) StrategyFactory {
	return StrategyFactory{
		Name:      name,
		Family:    family,
		Benchmark: "flat_cash",
		New: func() backtest.PortfolioStrategy {
			return &toggleStrategy{name: name, daily: daily}
		},
	}
}

func warmupToggleFactory(lookback int) StrategyFactory {
	return StrategyFactory{
		Name:      "warmup_toggle",
		Family:    "custom",
		Benchmark: "flat_cash",
		New: func() backtest.PortfolioStrategy {
			return &warmupToggleStrategy{lookback: lookback}
		},
	}
}

func metadataFactory() StrategyFactory {
	return StrategyFactory{
		Name:      "metadata_strategy",
		Family:    "custom",
		Benchmark: "flat_cash",
		New: func() backtest.PortfolioStrategy {
			return &metadataStrategy{}
		},
	}
}

type toggleStrategy struct {
	name  string
	daily bool
}

func (s *toggleStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	outputs := make([]backtest.PortfolioOutput, len(panel.Times))
	for i := range panel.Times {
		output, err := s.EvaluatePortfolioLatest(ctx, panelPrefix(panel, i+1))
		if err != nil {
			return nil, err
		}
		outputs[i] = output
	}
	return outputs, nil
}

func (s *toggleStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	_ = ctx
	i := len(panel.Times) - 1
	t := panel.Times[i]
	weight := 1.0
	if s.daily && i%2 == 1 {
		weight = 0
	}
	return backtest.PortfolioOutput{
		Time: t,
		Targets: map[string]backtest.TargetPosition{
			"VOO": {
				Symbol:       "VOO",
				TargetWeight: weight,
				AlphaScore:   weight,
				Confidence:   1,
				Side:         backtest.PositionSideLong,
				Engine:       s.name,
			},
		},
		GrossExposure: weight,
		NetExposure:   weight,
		CashWeight:    1 - weight,
	}, nil
}

func (s *toggleStrategy) Universe() []string { return []string{"VOO"} }
func (s *toggleStrategy) Name() string       { return s.name }

type warmupToggleStrategy struct {
	lookback int
}

func (s *warmupToggleStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	outputs := make([]backtest.PortfolioOutput, len(panel.Times))
	for i := range panel.Times {
		output, err := s.EvaluatePortfolioLatest(ctx, panelPrefix(panel, i+1))
		if err != nil {
			return nil, err
		}
		outputs[i] = output
	}
	return outputs, nil
}

func (s *warmupToggleStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	_ = ctx
	i := len(panel.Times) - 1
	t := panel.Times[i]
	if i < s.lookback {
		return holdPortfolioOutput(t, "warmup"), nil
	}
	weight := 1.0
	if i%2 == 1 {
		weight = 0
	}
	return backtest.PortfolioOutput{
		Time: t,
		Targets: map[string]backtest.TargetPosition{
			"VOO": {
				Symbol:       "VOO",
				TargetWeight: weight,
				AlphaScore:   weight,
				Confidence:   1,
				Side:         backtest.PositionSideLong,
				Engine:       "warmup_toggle",
			},
		},
		GrossExposure: weight,
		NetExposure:   weight,
		CashWeight:    1 - weight,
	}, nil
}

func (s *warmupToggleStrategy) Universe() []string { return []string{"VOO"} }
func (s *warmupToggleStrategy) Name() string       { return "warmup_toggle" }

type metadataStrategy struct{}

func (s *metadataStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	outputs := make([]backtest.PortfolioOutput, len(panel.Times))
	for i := range panel.Times {
		outputs[i], _ = s.EvaluatePortfolioLatest(ctx, panelPrefix(panel, i+1))
	}
	return outputs, nil
}

func (s *metadataStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	_ = ctx
	t := panel.Times[len(panel.Times)-1]
	return backtest.PortfolioOutput{
		Time: t,
		EngineMetadata: map[string]interface{}{
			"engine":               "metadata_strategy",
			"ranker_model_loaded":  true,
			"ranker_model_sha256":  "abc123",
			"ranker_model_variant": "unit",
			"selection_rows":       []map[string]interface{}{{"symbol": "VOO"}},
		},
	}, nil
}

func (s *metadataStrategy) Universe() []string { return []string{"VOO"} }
func (s *metadataStrategy) Name() string       { return "metadata_strategy" }

func validationTestPanel(n int) backtest.AlignedBars {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	times := make([]time.Time, n)
	bars := make([]models.Bar, n)
	for i := range bars {
		price := 100 + float64(i)*0.25
		times[i] = start.AddDate(0, 0, i)
		bars[i] = models.Bar{
			Time:   times[i],
			Symbol: "VOO",
			Open:   price,
			High:   price + 1,
			Low:    price - 1,
			Close:  price,
			Volume: 1_000_000,
		}
	}
	return backtest.AlignedBars{
		Times:     times,
		Symbols:   []string{"VOO"},
		Bars:      map[string][]models.Bar{"VOO": bars},
		Timeframe: "1Day",
	}
}

func containsReason(reasons []string, needle string) bool {
	for _, reason := range reasons {
		if strings.Contains(reason, needle) {
			return true
		}
	}
	return false
=======
	"testing"
	"time"

	"github.com/oalpha/pkg/models"
)

type stubSource struct {
	bars map[string][]models.Bar
}

func (s stubSource) LoadBars(_ context.Context, symbols []string, timeframe string, window ValidationWindow) (map[string][]models.Bar, error) {
	return s.bars, nil
}

func TestRunValidationProducesCandidates(t *testing.T) {
	bars := make(map[string][]models.Bar)
	bars["VOO"] = syntheticBars("VOO", 900, 100, 0.0004)
	bars["AAPL"] = syntheticBars("AAPL", 900, 50, 0.0009)
	bars["MSFT"] = syntheticBars("MSFT", 900, 60, 0.0008)
	bars["NVDA"] = syntheticBars("NVDA", 900, 70, 0.0011)
	window := ValidationWindow{From: bars["VOO"][0].Time, To: bars["VOO"][len(bars["VOO"])-1].Time}
	report, err := RunValidation(context.Background(), stubSource{bars: bars}, "benchmark_ranker_proxy_h63", RunnerConfig{
		Symbols:   []string{"VOO", "AAPL", "MSFT", "NVDA"},
		Timeframe: "1Day",
		Window:    window,
		TrainBars: 252,
		TestBars:  126,
		StepBars:  126,
		MinTrades: 1,
	})
	if err != nil {
		t.Fatalf("run validation: %v", err)
	}
	if len(report.Candidates) < 3 {
		t.Fatalf("expected sibling candidates, got %d", len(report.Candidates))
	}
	if report.Candidates[0].Strategy != "benchmark_ranker_proxy_h63_checkpoint" {
		t.Fatalf("expected primary variant first, got %s", report.Candidates[0].Strategy)
	}
}

func syntheticBars(symbol string, count int, startPrice, dailyDrift float64) []models.Bar {
	bars := make([]models.Bar, count)
	price := startPrice
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < count; i++ {
		price = price * (1 + dailyDrift + 0.01*float64((i%7)-3)/100)
		bars[i] = models.Bar{
			Time:   base.AddDate(0, 0, i),
			Symbol: symbol,
			Open:   price,
			High:   price * 1.01,
			Low:    price * 0.99,
			Close:  price,
			Volume: 1000,
		}
	}
	return bars
>>>>>>> 3ea6d428 (Alpha research)
}
