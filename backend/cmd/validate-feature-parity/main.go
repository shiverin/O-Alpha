package main

import (
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

	"gopkg.in/yaml.v3"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/ml"
	"github.com/oalpha/pkg/models"
)

type parityReport struct {
	GeneratedAt    time.Time              `json:"generated_at"`
	RowsChecked    int                    `json:"rows_checked"`
	FeatureCount   int                    `json:"feature_count"`
	Tolerance      float64                `json:"tolerance"`
	MaxAbsError    float64                `json:"max_abs_error"`
	Failures       int                    `json:"failures"`
	MissingRows    int                    `json:"missing_rows"`
	PerFeature     []featureParitySummary `json:"per_feature"`
	Fixture        string                 `json:"fixture"`
	BarsCSV        string                 `json:"bars_csv"`
	SignalsCSV     string                 `json:"signals_csv"`
	FeatureSpec    string                 `json:"feature_spec"`
	FeatureSpecVer string                 `json:"feature_spec_version"`
	FailureCSV     string                 `json:"failure_csv,omitempty"`
	Status         string                 `json:"status"`
	Diagnostics    map[string]interface{} `json:"diagnostics,omitempty"`
}

type featureParitySummary struct {
	Feature     string  `json:"feature"`
	MaxAbsError float64 `json:"max_abs_error"`
	MeanAbsErr  float64 `json:"mean_abs_error"`
	Failures    int     `json:"failures"`
}

type failureRow struct {
	Row      int
	Symbol   string
	Time     time.Time
	Feature  string
	Expected float64
	Actual   float64
	AbsError float64
}

func main() {
	barsPath := flag.String("bars-csv", "", "bars.csv used to build Python features")
	signalsPath := flag.String("signals-csv", "", "signals.csv used to build Python features")
	fixturePath := flag.String("fixture", "", "feature fixture CSV from research/ml/export_feature_fixture.py")
	featureSpecPath := flag.String("feature-spec", "", "feature spec YAML or JSON")
	outPath := flag.String("out", "", "JSON report path")
	failureCSVPath := flag.String("failure-csv", "", "optional per-feature failure CSV")
	tolerance := flag.Float64("tolerance", 1e-9, "absolute feature tolerance")
	flag.Parse()

	if *barsPath == "" || *signalsPath == "" || *fixturePath == "" || *featureSpecPath == "" {
		fail("usage: validate-feature-parity --bars-csv bars.csv --signals-csv signals.csv --fixture feature_fixture.csv --feature-spec feature_spec.yaml [--out report.json]")
	}
	if *outPath == "" {
		*outPath = strings.TrimSuffix(*fixturePath, filepath.Ext(*fixturePath)) + "_go_parity.json"
	}

	spec, err := readFeatureSpec(*featureSpecPath)
	if err != nil {
		fail("read feature spec: %v", err)
	}
	barsBySymbol, err := readBarsCSV(*barsPath)
	if err != nil {
		fail("read bars: %v", err)
	}
	signalsByKey, err := readSignalsCSV(*signalsPath)
	if err != nil {
		fail("read signals: %v", err)
	}
	fixture, header, err := readCSV(*fixturePath)
	if err != nil {
		fail("read fixture: %v", err)
	}
	featureIndexes, err := featureIndexes(header, spec.Features)
	if err != nil {
		fail("fixture columns: %v", err)
	}
	symbolIndex := columnIndex(header, "symbol")
	timeIndex := columnIndex(header, "event_time")
	if symbolIndex < 0 || timeIndex < 0 {
		fail("fixture must contain symbol and event_time")
	}

	contextBars := cloneContextBars(barsBySymbol)
	builder := ml.NewFeatureBuilder(spec)
	acc := make(map[string]*featureAccumulator, len(spec.Features))
	for _, feature := range spec.Features {
		acc[feature] = &featureAccumulator{}
	}

	var failures []failureRow
	var maxAbsError float64
	var missingRows int
	rowsChecked := 0
	for rowNumber, row := range fixture {
		symbol := strings.ToUpper(strings.TrimSpace(row[symbolIndex]))
		eventTime, err := parseTime(row[timeIndex])
		if err != nil {
			fail("parse fixture time row %d: %v", rowNumber, err)
		}
		bars := barsBySymbol[symbol]
		index := barIndexAt(bars, eventTime)
		if index < 0 {
			missingRows++
			continue
		}
		baseOutput := signalOutput(signalsByKey[barKey(symbol, eventTime)])
		vector, err := builder.BuildAt(ml.FeatureBuildInput{
			Symbol:      symbol,
			Bars:        bars,
			BaseOutput:  &baseOutput,
			ContextBars: contextBars,
		}, index)
		if err != nil {
			fail("build Go features row %d %s %s: %v", rowNumber, symbol, eventTime.Format(time.RFC3339), err)
		}
		rowsChecked++
		for i, feature := range spec.Features {
			expected, err := strconv.ParseFloat(row[featureIndexes[i]], 64)
			if err != nil {
				fail("parse expected row %d feature %s: %v", rowNumber, feature, err)
			}
			actual := vector.Values[i]
			absError := math.Abs(expected - actual)
			if absError > maxAbsError {
				maxAbsError = absError
			}
			acc[feature].observe(absError, absError > *tolerance)
			if absError > *tolerance {
				failures = append(failures, failureRow{
					Row:      rowNumber,
					Symbol:   symbol,
					Time:     eventTime,
					Feature:  feature,
					Expected: expected,
					Actual:   actual,
					AbsError: absError,
				})
			}
		}
	}

	perFeature := summarizeFeatures(spec.Features, acc)
	report := parityReport{
		GeneratedAt:    time.Now().UTC(),
		RowsChecked:    rowsChecked,
		FeatureCount:   len(spec.Features),
		Tolerance:      *tolerance,
		MaxAbsError:    maxAbsError,
		Failures:       len(failures),
		MissingRows:    missingRows,
		PerFeature:     perFeature,
		Fixture:        *fixturePath,
		BarsCSV:        *barsPath,
		SignalsCSV:     *signalsPath,
		FeatureSpec:    *featureSpecPath,
		FeatureSpecVer: spec.Version,
		FailureCSV:     *failureCSVPath,
		Status:         "passed",
		Diagnostics: map[string]interface{}{
			"symbols_loaded": len(barsBySymbol),
		},
	}
	if len(failures) > 0 || missingRows > 0 {
		report.Status = "failed"
	}
	if *failureCSVPath != "" {
		if err := writeFailureCSV(*failureCSVPath, failures); err != nil {
			fail("write failure CSV: %v", err)
		}
	}
	if err := writeJSON(*outPath, report); err != nil {
		fail("write report: %v", err)
	}
	fmt.Printf("feature parity %s rows=%d features=%d max_abs_error=%.12g failures=%d missing_rows=%d report=%s\n",
		report.Status, rowsChecked, len(spec.Features), maxAbsError, len(failures), missingRows, *outPath)
	if report.Status != "passed" {
		os.Exit(1)
	}
}

type featureAccumulator struct {
	count     int
	failures  int
	sumAbsErr float64
	maxAbsErr float64
}

func (a *featureAccumulator) observe(absErr float64, failed bool) {
	a.count++
	a.sumAbsErr += absErr
	if absErr > a.maxAbsErr {
		a.maxAbsErr = absErr
	}
	if failed {
		a.failures++
	}
}

func summarizeFeatures(features []string, acc map[string]*featureAccumulator) []featureParitySummary {
	out := make([]featureParitySummary, 0, len(features))
	for _, feature := range features {
		a := acc[feature]
		mean := 0.0
		if a.count > 0 {
			mean = a.sumAbsErr / float64(a.count)
		}
		out = append(out, featureParitySummary{
			Feature:     feature,
			MaxAbsError: a.maxAbsErr,
			MeanAbsErr:  mean,
			Failures:    a.failures,
		})
	}
	return out
}

func readFeatureSpec(path string) (ml.FeatureSpec, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return ml.FeatureSpec{}, err
	}
	var spec ml.FeatureSpec
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(payload, &spec)
	default:
		err = json.Unmarshal(payload, &spec)
	}
	if err != nil {
		return ml.FeatureSpec{}, err
	}
	if len(spec.Features) == 0 {
		return ml.FeatureSpec{}, fmt.Errorf("feature spec has no features")
	}
	return spec, nil
}

func readBarsCSV(path string) (map[string][]models.Bar, error) {
	rows, header, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	idx := requiredColumns(header, []string{"symbol", "time", "open", "high", "low", "close", "volume"})
	out := make(map[string][]models.Bar)
	for rowNumber, row := range rows {
		t, err := parseTime(row[idx["time"]])
		if err != nil {
			return nil, fmt.Errorf("row %d time: %w", rowNumber, err)
		}
		symbol := strings.ToUpper(strings.TrimSpace(row[idx["symbol"]]))
		open, err := parseFloat(row[idx["open"]])
		if err != nil {
			return nil, fmt.Errorf("row %d open: %w", rowNumber, err)
		}
		high, err := parseFloat(row[idx["high"]])
		if err != nil {
			return nil, fmt.Errorf("row %d high: %w", rowNumber, err)
		}
		low, err := parseFloat(row[idx["low"]])
		if err != nil {
			return nil, fmt.Errorf("row %d low: %w", rowNumber, err)
		}
		closePrice, err := parseFloat(row[idx["close"]])
		if err != nil {
			return nil, fmt.Errorf("row %d close: %w", rowNumber, err)
		}
		volumeFloat, err := parseFloat(row[idx["volume"]])
		if err != nil {
			return nil, fmt.Errorf("row %d volume: %w", rowNumber, err)
		}
		out[symbol] = append(out[symbol], models.Bar{
			Time:   t,
			Symbol: symbol,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closePrice,
			Volume: int64(math.Round(volumeFloat)),
		})
	}
	for symbol := range out {
		sort.Slice(out[symbol], func(i, j int) bool {
			return out[symbol][i].Time.Before(out[symbol][j].Time)
		})
	}
	return out, nil
}

type signalRecord struct {
	Signal       string
	AlphaScore   float64
	Confidence   float64
	TargetWeight float64
}

func readSignalsCSV(path string) (map[string]signalRecord, error) {
	rows, header, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	idx := requiredColumns(header, []string{"symbol", "time", "signal", "alpha_score", "confidence", "target_weight"})
	out := make(map[string]signalRecord, len(rows))
	for rowNumber, row := range rows {
		t, err := parseTime(row[idx["time"]])
		if err != nil {
			return nil, fmt.Errorf("row %d time: %w", rowNumber, err)
		}
		alphaScore, err := parseFloat(row[idx["alpha_score"]])
		if err != nil {
			return nil, fmt.Errorf("row %d alpha_score: %w", rowNumber, err)
		}
		confidence, err := parseFloat(row[idx["confidence"]])
		if err != nil {
			return nil, fmt.Errorf("row %d confidence: %w", rowNumber, err)
		}
		targetWeight, err := parseFloat(row[idx["target_weight"]])
		if err != nil {
			return nil, fmt.Errorf("row %d target_weight: %w", rowNumber, err)
		}
		symbol := strings.ToUpper(strings.TrimSpace(row[idx["symbol"]]))
		out[barKey(symbol, t)] = signalRecord{
			Signal:       row[idx["signal"]],
			AlphaScore:   alphaScore,
			Confidence:   confidence,
			TargetWeight: targetWeight,
		}
	}
	return out, nil
}

func signalOutput(record signalRecord) backtest.StrategyOutput {
	return backtest.StrategyOutput{
		Signal:       signalValue(record.Signal),
		AlphaScore:   record.AlphaScore,
		Confidence:   record.Confidence,
		TargetWeight: record.TargetWeight,
	}
}

func signalValue(value string) models.Signal {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "BUY", "LONG", "1":
		return models.SignalBuy
	case "SELL", "SHORT", "-1":
		return models.SignalSell
	default:
		return models.SignalHold
	}
}

func readCSV(path string) ([][]string, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = file.Close() }()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(records) == 0 {
		return nil, nil, fmt.Errorf("empty CSV")
	}
	return records[1:], records[0], nil
}

func featureIndexes(header []string, features []string) ([]int, error) {
	indexes := make([]int, len(features))
	for i, feature := range features {
		idx := columnIndex(header, feature)
		if idx < 0 {
			return nil, fmt.Errorf("missing feature %s", feature)
		}
		indexes[i] = idx
	}
	return indexes, nil
}

func requiredColumns(header []string, names []string) map[string]int {
	out := make(map[string]int, len(names))
	for _, name := range names {
		idx := columnIndex(header, name)
		if idx < 0 {
			fail("missing required CSV column %s", name)
		}
		out[name] = idx
	}
	return out
}

func columnIndex(header []string, name string) int {
	for i, column := range header {
		if column == name {
			return i
		}
	}
	return -1
}

func barIndexAt(bars []models.Bar, t time.Time) int {
	for i := range bars {
		if bars[i].Time.Equal(t) {
			return i
		}
	}
	return -1
}

func cloneContextBars(values map[string][]models.Bar) map[string][]models.Bar {
	out := make(map[string][]models.Bar, len(values))
	for symbol, bars := range values {
		out[symbol] = bars
	}
	return out
}

func parseTime(value string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err == nil {
		return t.UTC(), nil
	}
	return time.Parse("2006-01-02 15:04:05-07:00", strings.TrimSpace(value))
}

func parseFloat(value string) (float64, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	return strconv.ParseFloat(value, 64)
}

func barKey(symbol string, t time.Time) string {
	return strings.ToUpper(strings.TrimSpace(symbol)) + "|" + t.UTC().Format(time.RFC3339)
}

func writeFailureCSV(path string, rows []failureRow) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"row", "symbol", "time", "feature", "expected", "actual", "abs_error"}); err != nil {
		return err
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].AbsError == rows[j].AbsError {
			return rows[i].Feature < rows[j].Feature
		}
		return rows[i].AbsError > rows[j].AbsError
	})
	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.Itoa(row.Row),
			row.Symbol,
			row.Time.UTC().Format(time.RFC3339),
			row.Feature,
			strconv.FormatFloat(row.Expected, 'g', -1, 64),
			strconv.FormatFloat(row.Actual, 'g', -1, 64),
			strconv.FormatFloat(row.AbsError, 'g', -1, 64),
		}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func writeJSON(path string, value interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(payload, '\n'), 0o644)
}

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
