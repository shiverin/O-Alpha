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

	"github.com/oalpha/internal/ml"
	"github.com/oalpha/pkg/models"
)

type parityReport struct {
	GeneratedAt  time.Time              `json:"generated_at"`
	Status       string                 `json:"status"`
	RowsChecked  int                    `json:"rows_checked"`
	FeatureCount int                    `json:"feature_count"`
	Tolerance    float64                `json:"tolerance"`
	MaxAbsError  float64                `json:"max_abs_error"`
	Failures     int                    `json:"failures"`
	MissingRows  int                    `json:"missing_rows"`
	PerFeature   []featureParitySummary `json:"per_feature"`
	BarsCSV      string                 `json:"bars_csv"`
	Fixture      string                 `json:"fixture"`
	Benchmark    string                 `json:"benchmark"`
	FailureCSV   string                 `json:"failure_csv,omitempty"`
}

type featureParitySummary struct {
	Feature     string  `json:"feature"`
	MaxAbsError float64 `json:"max_abs_error"`
	MeanAbsErr  float64 `json:"mean_abs_error"`
	Failures    int     `json:"failures"`
}

type featureAccumulator struct {
	count     int
	failures  int
	sumAbsErr float64
	maxAbsErr float64
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
	barsPath := flag.String("bars-csv", "", "bars.csv used to build Python daily ranker features")
	fixturePath := flag.String("fixture", "", "CSV from research/ml/export_daily_ranker_feature_fixture.py")
	benchmark := flag.String("benchmark", "VOO", "benchmark symbol")
	outPath := flag.String("out", "", "JSON report path")
	failureCSVPath := flag.String("failure-csv", "", "optional failure CSV")
	tolerance := flag.Float64("tolerance", 1e-8, "absolute feature tolerance")
	flag.Parse()

	if *barsPath == "" || *fixturePath == "" {
		fail("usage: validate-daily-ranker-parity --bars-csv bars.csv --fixture fixture.csv [--out report.json]")
	}
	if *outPath == "" {
		*outPath = strings.TrimSuffix(*fixturePath, filepath.Ext(*fixturePath)) + "_go_parity.json"
	}

	barsBySymbol, err := readBarsCSV(*barsPath)
	if err != nil {
		fail("read bars: %v", err)
	}
	fixture, header, err := readCSV(*fixturePath)
	if err != nil {
		fail("read fixture: %v", err)
	}
	symbolIndex := columnIndex(header, "symbol")
	timeIndex := columnIndex(header, "event_time")
	if symbolIndex < 0 || timeIndex < 0 {
		fail("fixture must contain symbol and event_time")
	}
	featureIndexes, err := featureIndexes(header, ml.DailyRankerFeatureNames)
	if err != nil {
		fail("fixture feature columns: %v", err)
	}

	builder := ml.NewDailyRankerFeatureBuilder(*benchmark)
	acc := make(map[string]*featureAccumulator, len(ml.DailyRankerFeatureNames))
	for _, feature := range ml.DailyRankerFeatureNames {
		acc[feature] = &featureAccumulator{}
	}

	var failures []failureRow
	var maxAbsError float64
	var missingRows int
	var rowsChecked int
	for rowNumber, row := range fixture {
		symbol := strings.ToUpper(strings.TrimSpace(row[symbolIndex]))
		eventTime, err := parseTime(row[timeIndex])
		if err != nil {
			fail("parse fixture time row %d: %v", rowNumber, err)
		}
		vector, err := builder.BuildAtTime(barsBySymbol, symbol, eventTime)
		if err != nil {
			missingRows++
			continue
		}
		rowsChecked++
		for i, feature := range ml.DailyRankerFeatureNames {
			expected, err := strconv.ParseFloat(row[featureIndexes[i]], 64)
			if err != nil {
				fail("parse expected row %d feature %s: %v", rowNumber, feature, err)
			}
			actual := vector.Values[i]
			absError := math.Abs(expected - actual)
			if absError > maxAbsError {
				maxAbsError = absError
			}
			failed := absError > *tolerance
			acc[feature].observe(absError, failed)
			if failed {
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

	report := parityReport{
		GeneratedAt:  time.Now().UTC(),
		Status:       "passed",
		RowsChecked:  rowsChecked,
		FeatureCount: len(ml.DailyRankerFeatureNames),
		Tolerance:    *tolerance,
		MaxAbsError:  maxAbsError,
		Failures:     len(failures),
		MissingRows:  missingRows,
		PerFeature:   summarizeFeatures(ml.DailyRankerFeatureNames, acc),
		BarsCSV:      *barsPath,
		Fixture:      *fixturePath,
		Benchmark:    strings.ToUpper(strings.TrimSpace(*benchmark)),
		FailureCSV:   *failureCSVPath,
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
	fmt.Printf("daily ranker feature parity %s rows=%d features=%d max_abs_error=%.12g failures=%d missing_rows=%d report=%s\n",
		report.Status, rowsChecked, len(ml.DailyRankerFeatureNames), maxAbsError, len(failures), missingRows, *outPath)
	if report.Status != "passed" {
		os.Exit(1)
	}
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
		openPrice, err := parseFloat(row[idx["open"]])
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
		volume, err := parseFloat(row[idx["volume"]])
		if err != nil {
			return nil, fmt.Errorf("row %d volume: %w", rowNumber, err)
		}
		symbol := strings.ToUpper(strings.TrimSpace(row[idx["symbol"]]))
		out[symbol] = append(out[symbol], models.Bar{
			Time:   t,
			Symbol: symbol,
			Open:   openPrice,
			High:   high,
			Low:    low,
			Close:  closePrice,
			Volume: int64(math.Round(volume)),
		})
	}
	for symbol := range out {
		sort.Slice(out[symbol], func(i, j int) bool {
			return out[symbol][i].Time.Before(out[symbol][j].Time)
		})
	}
	return out, nil
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

func parseTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05-07:00"} {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time %q", value)
}

func parseFloat(value string) (float64, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	return strconv.ParseFloat(value, 64)
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
