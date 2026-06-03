package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/oalpha/internal/alpha/ranker"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/research/panelload"
)

const (
	defaultModelRoot    = "../reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts"
	defaultModelVariant = "stocks_h63_s15_top3_reb63_z10"
	defaultYears        = "2018,2019,2020,2021,2022,2023,2024,2025,2026"
)

type signalArgs struct {
	BarsCSV                 string    `json:"bars_csv"`
	Symbols                 []string  `json:"symbols"`
	Timeframe               string    `json:"timeframe"`
	From                    time.Time `json:"from"`
	To                      time.Time `json:"to"`
	Benchmark               string    `json:"benchmark"`
	CandidateUniverse       string    `json:"candidate_universe"`
	ExcludedSymbols         []string  `json:"excluded_symbols,omitempty"`
	ModelArtifactRoot       string    `json:"model_artifact_root"`
	ModelVariant            string    `json:"model_variant"`
	ModelYears              []int     `json:"model_years"`
	PointInTimeUniversePath string    `json:"point_in_time_universe_path,omitempty"`
	RebalanceEveryBars      int       `json:"rebalance_every_bars"`
	SleeveFraction          float64   `json:"sleeve_fraction"`
	TopK                    int       `json:"top_k"`
	MaxNameWeight           float64   `json:"max_name_weight"`
	TurnoverBand            float64   `json:"turnover_band"`
	MinScoreZ               float64   `json:"min_score_z"`
	MaxCandidateVol         float64   `json:"max_candidate_vol,omitempty"`
	MaxBenchmarkVol         float64   `json:"max_benchmark_vol,omitempty"`
	HighVolScale            float64   `json:"high_vol_scale,omitempty"`
	MaxBenchmarkDrawdown    float64   `json:"max_benchmark_drawdown,omitempty"`
	DrawdownScale           float64   `json:"drawdown_scale,omitempty"`
}

type paperSignalReport struct {
	GeneratedAt             time.Time              `json:"generated_at"`
	PaperOnly               bool                   `json:"paper_only"`
	OrdersEnabled           bool                   `json:"orders_enabled"`
	OrdersSubmitted         int                    `json:"orders_submitted"`
	BrokerConnected         bool                   `json:"broker_connected"`
	Strategy                string                 `json:"strategy"`
	Args                    signalArgs             `json:"args"`
	PanelStart              time.Time              `json:"panel_start"`
	PanelEnd                time.Time              `json:"panel_end"`
	BarCount                int                    `json:"bar_count"`
	LatestTime              time.Time              `json:"latest_time"`
	LatestAction            string                 `json:"latest_action"`
	LatestMetadata          map[string]interface{} `json:"latest_metadata,omitempty"`
	LastRebalanceTime       *time.Time             `json:"last_rebalance_time,omitempty"`
	LastRebalanceMetadata   map[string]interface{} `json:"last_rebalance_metadata,omitempty"`
	TargetSource            string                 `json:"target_source"`
	Targets                 []targetRow            `json:"targets"`
	Warnings                []string               `json:"warnings"`
	LatestGrossExposure     float64                `json:"latest_gross_exposure"`
	LatestNetExposure       float64                `json:"latest_net_exposure"`
	LastTargetGrossExposure float64                `json:"last_target_gross_exposure"`
	LastTargetNetExposure   float64                `json:"last_target_net_exposure"`
}

type targetRow struct {
	Symbol       string                 `json:"symbol"`
	TargetWeight float64                `json:"target_weight"`
	AlphaScore   float64                `json:"alpha_score"`
	ModelScore   *float64               `json:"model_score,omitempty"`
	ScoreZ       *float64               `json:"score_z,omitempty"`
	Vol20        *float64               `json:"vol_20,omitempty"`
	Rank         *int                   `json:"rank,omitempty"`
	Confidence   float64                `json:"confidence"`
	Side         string                 `json:"side"`
	Engine       string                 `json:"engine"`
	Role         string                 `json:"role,omitempty"`
	Reason       string                 `json:"reason,omitempty"`
	Rebalance    bool                   `json:"rebalance"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type selectionRow struct {
	Rank   int
	Score  float64
	ScoreZ float64
	Vol20  float64
}

func main() {
	var (
		barsCSV              = flag.String("bars-csv", "", "exported bars CSV; required")
		symbols              = flag.String("symbols", "", "comma-separated symbols; include benchmark first")
		timeframe            = flag.String("timeframe", "1Day", "bar timeframe")
		from                 = flag.String("from", "2015-01-01", "inclusive start date, YYYY-MM-DD")
		to                   = flag.String("to", "", "inclusive end date, YYYY-MM-DD; defaults to now")
		benchmark            = flag.String("benchmark", "VOO", "benchmark/core symbol")
		candidateUniverse    = flag.String("candidate-universe", "stocks", "candidate universe filter: stocks, etfs, or all")
		excludedSymbols      = flag.String("excluded-symbols", "", "comma-separated active-sleeve symbols to exclude")
		modelRoot            = flag.String("model-root", defaultRankerModelRoot(), "daily ranker fold artifact root")
		modelVariant         = flag.String("variant", defaultModelVariant, "daily ranker model variant")
		modelYears           = flag.String("years", defaultYears, "comma-separated model years")
		pitUniverse          = flag.String("pit-universe", strings.TrimSpace(os.Getenv("OALPHA_DAILY_RANKER_PIT_UNIVERSE")), "optional point-in-time universe JSON")
		rebalanceEveryBars   = flag.Int("rebalance-bars", 63, "rebalance cadence in bars")
		sleeveFraction       = flag.Float64("sleeve", 0.15, "active sleeve fraction")
		topK                 = flag.Int("top-k", 3, "number of active names")
		maxNameWeight        = flag.Float64("max-name-weight", 0.05, "maximum single active-name weight")
		turnoverBand         = flag.Float64("turnover-band", 0.05, "minimum turnover needed to refresh targets")
		minScoreZ            = flag.Float64("min-score-z", 1.0, "minimum cross-sectional score z for active selection")
		maxCandidateVol      = flag.Float64("max-candidate-vol", 0, "optional candidate annualized vol cap")
		maxBenchmarkVol      = flag.Float64("max-benchmark-vol", 0, "optional benchmark annualized vol cap")
		highVolScale         = flag.Float64("high-vol-scale", 0, "active sleeve scale when benchmark vol exceeds cap")
		maxBenchmarkDrawdown = flag.Float64("max-benchmark-drawdown", 0, "optional benchmark drawdown cap")
		drawdownScale        = flag.Float64("drawdown-scale", 0, "active sleeve scale when benchmark drawdown exceeds cap")
		outputDir            = flag.String("output-dir", "", "output directory; defaults to reports batch folder")
		printSummary         = flag.Bool("summary", true, "print written artifact paths")
	)
	flag.Parse()

	args, err := buildArgs(
		*barsCSV,
		*symbols,
		*timeframe,
		*from,
		*to,
		*benchmark,
		*candidateUniverse,
		*excludedSymbols,
		*modelRoot,
		*modelVariant,
		*modelYears,
		*pitUniverse,
		*rebalanceEveryBars,
		*sleeveFraction,
		*topK,
		*maxNameWeight,
		*turnoverBand,
		*minScoreZ,
		*maxCandidateVol,
		*maxBenchmarkVol,
		*highVolScale,
		*maxBenchmarkDrawdown,
		*drawdownScale,
	)
	if err != nil {
		fatal(err)
	}

	panel, err := panelload.LoadPanelFromCSV(args.BarsCSV, args.Symbols, args.Timeframe, args.From, args.To)
	if err != nil {
		fatal(err)
	}
	panel.Symbols = panelload.OrderSymbols(panel.Symbols, args.Symbols)
	if len(panel.Times) == 0 {
		fatal(fmt.Errorf("no aligned bars found for %s", strings.Join(args.Symbols, ",")))
	}

	strategy := ranker.NewDailyRankerSleeveStrategy(panel.Symbols, ranker.DailyRankerSleeveConfig{
		BenchmarkSymbol:         args.Benchmark,
		CandidateUniverse:       args.CandidateUniverse,
		ExcludedSymbols:         args.ExcludedSymbols,
		PointInTimeUniversePath: args.PointInTimeUniversePath,
		ModelArtifactRoot:       args.ModelArtifactRoot,
		ModelVariant:            args.ModelVariant,
		ModelPathsByYear:        ranker.DailyRankerModelPaths(args.ModelArtifactRoot, args.ModelVariant, args.ModelYears...),
		RebalanceEveryBars:      args.RebalanceEveryBars,
		SleeveFraction:          args.SleeveFraction,
		TopK:                    args.TopK,
		MaxNameWeight:           args.MaxNameWeight,
		TurnoverBand:            args.TurnoverBand,
		MinScoreZ:               args.MinScoreZ,
		MaxCandidateVol:         args.MaxCandidateVol,
		MaxBenchmarkVol:         args.MaxBenchmarkVol,
		HighVolScale:            args.HighVolScale,
		MaxBenchmarkDrawdown:    args.MaxBenchmarkDrawdown,
		DrawdownScale:           args.DrawdownScale,
	})
	outputs, err := strategy.GeneratePortfolioSignals(context.Background(), panel)
	if err != nil {
		fatal(err)
	}
	report, err := buildReport(args, panel, outputs, time.Now().UTC())
	if err != nil {
		fatal(err)
	}

	dir := strings.TrimSpace(*outputDir)
	if dir == "" {
		dir = filepath.Join("..", "reports", "batches", time.Now().UTC().Format("2006-01-02")+"_paper_ranker_signal")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fatal(err)
	}
	jsonPath := filepath.Join(dir, "paper_ranker_signal.json")
	csvPath := filepath.Join(dir, "paper_ranker_targets.csv")
	mdPath := filepath.Join(dir, "paper_ranker_signal.md")
	if err := writeJSON(jsonPath, report); err != nil {
		fatal(err)
	}
	if err := writeTargetsCSV(csvPath, report.Targets); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(mdPath, []byte(report.Markdown()), 0o644); err != nil {
		fatal(err)
	}
	if *printSummary {
		fmt.Printf("paper ranker signal written:\n  %s\n  %s\n  %s\n", jsonPath, csvPath, mdPath)
		fmt.Printf("paper_only=%t orders_enabled=%t orders_submitted=%d broker_connected=%t latest=%s targets=%d source=%s\n",
			report.PaperOnly,
			report.OrdersEnabled,
			report.OrdersSubmitted,
			report.BrokerConnected,
			report.LatestTime.Format(time.RFC3339),
			len(report.Targets),
			report.TargetSource,
		)
	}
}

func buildArgs(
	barsCSV string,
	symbolsCSV string,
	timeframe string,
	from string,
	to string,
	benchmark string,
	candidateUniverse string,
	excludedSymbolsCSV string,
	modelRoot string,
	modelVariant string,
	modelYearsCSV string,
	pitUniverse string,
	rebalanceEveryBars int,
	sleeveFraction float64,
	topK int,
	maxNameWeight float64,
	turnoverBand float64,
	minScoreZ float64,
	maxCandidateVol float64,
	maxBenchmarkVol float64,
	highVolScale float64,
	maxBenchmarkDrawdown float64,
	drawdownScale float64,
) (signalArgs, error) {
	if strings.TrimSpace(barsCSV) == "" {
		return signalArgs{}, fmt.Errorf("-bars-csv is required")
	}
	symbols := panelload.ParseCSVSymbols(symbolsCSV)
	if len(symbols) == 0 {
		return signalArgs{}, fmt.Errorf("-symbols is required")
	}
	benchmark = strings.ToUpper(strings.TrimSpace(benchmark))
	if benchmark == "" {
		benchmark = symbols[0]
	}
	if !containsSymbol(symbols, benchmark) {
		symbols = append([]string{benchmark}, symbols...)
	}
	start, err := panelload.ParseDate(from, false)
	if err != nil {
		return signalArgs{}, err
	}
	var end time.Time
	if strings.TrimSpace(to) == "" {
		end = time.Now().UTC()
	} else {
		end, err = panelload.ParseDate(to, true)
		if err != nil {
			return signalArgs{}, err
		}
	}
	if !end.After(start) {
		return signalArgs{}, fmt.Errorf("to must be after from")
	}
	years, err := parseYears(modelYearsCSV)
	if err != nil {
		return signalArgs{}, err
	}
	if strings.TrimSpace(modelRoot) == "" {
		return signalArgs{}, fmt.Errorf("-model-root is required")
	}
	if strings.TrimSpace(modelVariant) == "" {
		return signalArgs{}, fmt.Errorf("-variant is required")
	}
	if rebalanceEveryBars <= 0 {
		return signalArgs{}, fmt.Errorf("-rebalance-bars must be positive")
	}
	if topK <= 0 {
		return signalArgs{}, fmt.Errorf("-top-k must be positive")
	}
	if !finiteNonNegative(sleeveFraction) || sleeveFraction > 1 {
		return signalArgs{}, fmt.Errorf("-sleeve must be between 0 and 1")
	}
	if !finiteNonNegative(maxNameWeight) || maxNameWeight > 1 {
		return signalArgs{}, fmt.Errorf("-max-name-weight must be between 0 and 1")
	}
	if !finiteNonNegative(turnoverBand) || turnoverBand > 1 {
		return signalArgs{}, fmt.Errorf("-turnover-band must be between 0 and 1")
	}
	return signalArgs{
		BarsCSV:                 strings.TrimSpace(barsCSV),
		Symbols:                 symbols,
		Timeframe:               strings.TrimSpace(timeframe),
		From:                    start,
		To:                      end,
		Benchmark:               benchmark,
		CandidateUniverse:       strings.ToLower(strings.TrimSpace(candidateUniverse)),
		ExcludedSymbols:         panelload.ParseCSVSymbols(excludedSymbolsCSV),
		ModelArtifactRoot:       filepath.Clean(strings.TrimSpace(modelRoot)),
		ModelVariant:            strings.TrimSpace(modelVariant),
		ModelYears:              years,
		PointInTimeUniversePath: strings.TrimSpace(pitUniverse),
		RebalanceEveryBars:      rebalanceEveryBars,
		SleeveFraction:          sleeveFraction,
		TopK:                    topK,
		MaxNameWeight:           maxNameWeight,
		TurnoverBand:            turnoverBand,
		MinScoreZ:               minScoreZ,
		MaxCandidateVol:         maxCandidateVol,
		MaxBenchmarkVol:         maxBenchmarkVol,
		HighVolScale:            highVolScale,
		MaxBenchmarkDrawdown:    maxBenchmarkDrawdown,
		DrawdownScale:           drawdownScale,
	}, nil
}

func buildReport(args signalArgs, panel backtest.AlignedBars, outputs []backtest.PortfolioOutput, generatedAt time.Time) (paperSignalReport, error) {
	if len(panel.Times) == 0 {
		return paperSignalReport{}, fmt.Errorf("panel has no bars")
	}
	if len(outputs) == 0 {
		return paperSignalReport{}, fmt.Errorf("strategy produced no outputs")
	}
	latest := outputs[len(outputs)-1]
	targetOutput := latest
	targetSource := "latest"
	var lastRebalanceTime *time.Time
	for i := len(outputs) - 1; i >= 0; i-- {
		if len(outputs[i].Targets) == 0 {
			continue
		}
		targetOutput = outputs[i]
		targetSource = "last_non_empty_target"
		t := outputs[i].Time
		lastRebalanceTime = &t
		break
	}
	if targetSource == "latest" && len(latest.Targets) > 0 {
		t := latest.Time
		lastRebalanceTime = &t
	}
	return paperSignalReport{
		GeneratedAt:             generatedAt,
		PaperOnly:               true,
		OrdersEnabled:           false,
		OrdersSubmitted:         0,
		BrokerConnected:         false,
		Strategy:                ranker.DailyRankerSleeveStrategyName,
		Args:                    args,
		PanelStart:              panel.Times[0],
		PanelEnd:                panel.Times[len(panel.Times)-1],
		BarCount:                len(panel.Times),
		LatestTime:              latest.Time,
		LatestAction:            metadataString(latest.EngineMetadata, "action"),
		LatestMetadata:          latest.EngineMetadata,
		LastRebalanceTime:       lastRebalanceTime,
		LastRebalanceMetadata:   targetOutput.EngineMetadata,
		TargetSource:            targetSource,
		Targets:                 sortedTargetRows(targetOutput.Targets, args.Benchmark, targetOutput.EngineMetadata),
		Warnings:                reportWarnings(args),
		LatestGrossExposure:     latest.GrossExposure,
		LatestNetExposure:       latest.NetExposure,
		LastTargetGrossExposure: targetOutput.GrossExposure,
		LastTargetNetExposure:   targetOutput.NetExposure,
	}, nil
}

func sortedTargetRows(targets map[string]backtest.TargetPosition, benchmark string, metadata map[string]interface{}) []targetRow {
	selections := selectionRowsBySymbol(metadata)
	rows := make([]targetRow, 0, len(targets))
	for _, target := range targets {
		row := targetRow{
			Symbol:       target.Symbol,
			TargetWeight: target.TargetWeight,
			AlphaScore:   target.AlphaScore,
			Confidence:   target.Confidence,
			Side:         string(target.Side),
			Engine:       target.Engine,
			Role:         metadataString(target.Metadata, "role"),
			Reason:       metadataString(target.Metadata, "reason"),
			Rebalance:    metadataBool(target.Metadata, "rebalance"),
			Metadata:     target.Metadata,
		}
		if selection, ok := selections[target.Symbol]; ok {
			row.ModelScore = floatPtr(selection.Score)
			row.ScoreZ = floatPtr(selection.ScoreZ)
			row.Vol20 = floatPtr(selection.Vol20)
			row.Rank = intPtr(selection.Rank)
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		iBench := rows[i].Symbol == benchmark
		jBench := rows[j].Symbol == benchmark
		if iBench != jBench {
			return iBench
		}
		if rows[i].TargetWeight != rows[j].TargetWeight {
			return rows[i].TargetWeight > rows[j].TargetWeight
		}
		return rows[i].Symbol < rows[j].Symbol
	})
	return rows
}

func reportWarnings(args signalArgs) []string {
	warnings := []string{
		"research_simulation_only",
		"no_orders_submitted",
		"broker_client_not_loaded",
	}
	if strings.TrimSpace(args.PointInTimeUniversePath) == "" {
		warnings = append(warnings, "static_symbol_panel_external_validity_not_cleared")
	}
	return warnings
}

func (r paperSignalReport) Markdown() string {
	var b strings.Builder
	b.WriteString("# Paper Ranker Signal\n\n")
	fmt.Fprintf(&b, "- Generated at: `%s`\n", r.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "- Strategy: `%s`\n", r.Strategy)
	fmt.Fprintf(&b, "- Paper only: `%t`\n", r.PaperOnly)
	fmt.Fprintf(&b, "- Orders enabled: `%t`\n", r.OrdersEnabled)
	fmt.Fprintf(&b, "- Orders submitted: `%d`\n", r.OrdersSubmitted)
	fmt.Fprintf(&b, "- Broker connected: `%t`\n", r.BrokerConnected)
	fmt.Fprintf(&b, "- Panel: `%s` to `%s` (`%d` bars)\n", r.PanelStart.Format("2006-01-02"), r.PanelEnd.Format("2006-01-02"), r.BarCount)
	fmt.Fprintf(&b, "- Latest signal time: `%s`\n", r.LatestTime.Format(time.RFC3339))
	if r.LastRebalanceTime != nil {
		fmt.Fprintf(&b, "- Last target refresh: `%s`\n", r.LastRebalanceTime.Format(time.RFC3339))
	}
	fmt.Fprintf(&b, "- Target source: `%s`\n", r.TargetSource)
	fmt.Fprintf(&b, "- Model variant: `%s`\n", r.Args.ModelVariant)
	fmt.Fprintf(&b, "- Model root: `%s`\n", r.Args.ModelArtifactRoot)
	fmt.Fprintf(&b, "- Benchmark/core: `%s`\n\n", r.Args.Benchmark)

	if len(r.Warnings) > 0 {
		b.WriteString("## Warnings\n\n")
		for _, warning := range r.Warnings {
			fmt.Fprintf(&b, "- `%s`\n", warning)
		}
		b.WriteString("\n")
	}

	b.WriteString("## Targets\n\n")
	b.WriteString("| Symbol | Weight | Role | Rank | Model Score | Score Z | Vol 20 | Confidence | Reason |\n")
	b.WriteString("| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- |\n")
	for _, target := range r.Targets {
		fmt.Fprintf(&b, "| `%s` | %.6f | `%s` | %s | %s | %s | %s | %.3f | `%s` |\n",
			escapeTable(target.Symbol),
			target.TargetWeight,
			escapeTable(target.Role),
			formatIntPtr(target.Rank),
			formatFloatPtr(target.ModelScore),
			formatFloatPtr(target.ScoreZ),
			formatFloatPtr(target.Vol20),
			target.Confidence,
			escapeTable(target.Reason),
		)
	}
	b.WriteString("\n## Latest Metadata\n\n")
	writeMetadataMarkdown(&b, r.LatestMetadata)
	b.WriteString("\n## Last Target Metadata\n\n")
	writeMetadataMarkdown(&b, r.LastRebalanceMetadata)
	return b.String()
}

func writeMetadataMarkdown(b *strings.Builder, metadata map[string]interface{}) {
	if len(metadata) == 0 {
		b.WriteString("_none_\n")
		return
	}
	keys := make([]string, 0, len(metadata))
	for key := range metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	b.WriteString("| Key | Value |\n")
	b.WriteString("| --- | --- |\n")
	for _, key := range keys {
		value := metadata[key]
		if key == "selection_rows" {
			value = compactSelectionRows(value)
		}
		fmt.Fprintf(b, "| `%s` | `%s` |\n", escapeTable(key), escapeTable(formatMetadataValue(value)))
	}
}

func writeTargetsCSV(path string, targets []targetRow) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.Write([]string{"symbol", "target_weight", "role", "alpha_score", "model_score", "score_z", "vol_20", "rank", "confidence", "side", "engine", "reason", "rebalance"}); err != nil {
		return err
	}
	for _, target := range targets {
		if err := writer.Write([]string{
			target.Symbol,
			fmt.Sprintf("%.10f", target.TargetWeight),
			target.Role,
			fmt.Sprintf("%.10f", target.AlphaScore),
			formatFloatPtr(target.ModelScore),
			formatFloatPtr(target.ScoreZ),
			formatFloatPtr(target.Vol20),
			formatIntPtr(target.Rank),
			fmt.Sprintf("%.10f", target.Confidence),
			target.Side,
			target.Engine,
			target.Reason,
			strconv.FormatBool(target.Rebalance),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func writeJSON(path string, value interface{}) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(payload, '\n'), 0o644)
}

func parseYears(value string) ([]int, error) {
	parts := strings.Split(value, ",")
	out := make([]int, 0, len(parts))
	seen := make(map[int]bool, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		year, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("parse model year %q: %w", part, err)
		}
		if year < 1900 || year > 2200 {
			return nil, fmt.Errorf("model year %d outside expected range", year)
		}
		if seen[year] {
			continue
		}
		seen[year] = true
		out = append(out, year)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("at least one model year is required")
	}
	sort.Ints(out)
	return out, nil
}

func defaultRankerModelRoot() string {
	if root := strings.TrimSpace(os.Getenv("OALPHA_DAILY_RANKER_ARTIFACT_ROOT")); root != "" {
		return root
	}
	return defaultModelRoot
}

func containsSymbol(symbols []string, wanted string) bool {
	for _, symbol := range symbols {
		if strings.EqualFold(symbol, wanted) {
			return true
		}
	}
	return false
}

func metadataString(metadata map[string]interface{}, key string) string {
	if metadata == nil {
		return ""
	}
	switch value := metadata[key].(type) {
	case string:
		return value
	case fmt.Stringer:
		return value.String()
	default:
		if value == nil {
			return ""
		}
		return fmt.Sprint(value)
	}
}

func metadataBool(metadata map[string]interface{}, key string) bool {
	if metadata == nil {
		return false
	}
	switch value := metadata[key].(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(value, "true")
	default:
		return false
	}
}

func metadataFloat(metadata map[string]interface{}, key string) float64 {
	if metadata == nil {
		return 0
	}
	switch value := metadata[key].(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case string:
		out, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return out
		}
	}
	return 0
}

func selectionRowsBySymbol(metadata map[string]interface{}) map[string]selectionRow {
	out := make(map[string]selectionRow)
	if metadata == nil {
		return out
	}
	add := func(row map[string]interface{}) {
		symbol := strings.ToUpper(strings.TrimSpace(metadataString(row, "symbol")))
		if symbol == "" {
			return
		}
		out[symbol] = selectionRow{
			Rank:   int(metadataFloat(row, "rank")),
			Score:  metadataFloat(row, "score"),
			ScoreZ: metadataFloat(row, "score_z"),
			Vol20:  metadataFloat(row, "vol_20"),
		}
	}
	switch rows := metadata["selection_rows"].(type) {
	case []map[string]interface{}:
		for _, row := range rows {
			add(row)
		}
	case []interface{}:
		for _, raw := range rows {
			if row, ok := raw.(map[string]interface{}); ok {
				add(row)
			}
		}
	}
	return out
}

func floatPtr(value float64) *float64 {
	return &value
}

func intPtr(value int) *int {
	return &value
}

func formatFloatPtr(value *float64) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%.6f", *value)
}

func formatIntPtr(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

func formatMetadataValue(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case []string:
		return strings.Join(v, ",")
	default:
		payload, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(payload)
	}
}

func compactSelectionRows(value interface{}) interface{} {
	rows, ok := value.([]map[string]interface{})
	if ok {
		return fmt.Sprintf("%d rows", len(rows))
	}
	asInterface, ok := value.([]interface{})
	if ok {
		return fmt.Sprintf("%d rows", len(asInterface))
	}
	return value
}

func escapeTable(value string) string {
	return strings.ReplaceAll(value, "|", "\\|")
}

func finiteNonNegative(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
