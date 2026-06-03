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
)

type parityReport struct {
	GeneratedAt time.Time `json:"generated_at"`
	Status      string    `json:"status"`
	RowsChecked int       `json:"rows_checked"`
	Tolerance   float64   `json:"tolerance"`
	MaxAbsError float64   `json:"max_abs_error"`
	Failures    int       `json:"failures"`
	Model       string    `json:"model"`
	Fixture     string    `json:"fixture"`
	OutputCSV   string    `json:"output_csv,omitempty"`
}

type predictionRow struct {
	Row      int
	Expected float64
	Actual   float64
	AbsError float64
}

func main() {
	modelPath := flag.String("model", "", "daily ranker LightGBM model.txt")
	fixturePath := flag.String("fixture", "", "fixture CSV with daily ranker features and expected_score")
	outPath := flag.String("out", "", "JSON report path")
	outputCSVPath := flag.String("output-csv", "", "optional prediction comparison CSV")
	tolerance := flag.Float64("tolerance", 1e-8, "absolute raw-score tolerance")
	flag.Parse()

	if *modelPath == "" || *fixturePath == "" {
		fail("usage: validate-daily-ranker-leaves-parity --model model.txt --fixture fixture.csv [--out report.json]")
	}
	if *outPath == "" {
		*outPath = strings.TrimSuffix(*fixturePath, filepath.Ext(*fixturePath)) + "_leaves_parity.json"
	}

	predictor, err := ml.NewRawLeavesPredictor(*modelPath, ml.FeatureSpec{
		Version:  "daily_ranker_v1",
		Features: ml.DailyRankerFeatureNames,
	}, *modelPath)
	if err != nil {
		fail("load ranker model: %v", err)
	}

	rows, header, err := readCSV(*fixturePath)
	if err != nil {
		fail("read fixture: %v", err)
	}
	featureIndexes, err := featureColumnIndexes(header, ml.DailyRankerFeatureNames)
	if err != nil {
		fail("fixture feature columns: %v", err)
	}
	expectedIndex := columnIndex(header, "expected_score")
	if expectedIndex < 0 {
		fail("fixture must contain expected_score")
	}

	predictions := make([]predictionRow, 0, len(rows))
	var maxAbsError float64
	var failures int
	for i, row := range rows {
		features, err := parseFeatureRow(row, featureIndexes)
		if err != nil {
			fail("parse feature row %d: %v", i, err)
		}
		expected, err := strconv.ParseFloat(row[expectedIndex], 64)
		if err != nil {
			fail("parse expected_score row %d: %v", i, err)
		}
		actual, err := predictor.PredictRaw(features)
		if err != nil {
			fail("predict row %d: %v", i, err)
		}
		absError := math.Abs(expected - actual)
		if absError > maxAbsError {
			maxAbsError = absError
		}
		if absError > *tolerance {
			failures++
		}
		predictions = append(predictions, predictionRow{
			Row:      i,
			Expected: expected,
			Actual:   actual,
			AbsError: absError,
		})
	}

	report := parityReport{
		GeneratedAt: time.Now().UTC(),
		Status:      "passed",
		RowsChecked: len(predictions),
		Tolerance:   *tolerance,
		MaxAbsError: maxAbsError,
		Failures:    failures,
		Model:       *modelPath,
		Fixture:     *fixturePath,
		OutputCSV:   *outputCSVPath,
	}
	if failures > 0 {
		report.Status = "failed"
	}
	if *outputCSVPath != "" {
		if err := writePredictions(*outputCSVPath, predictions); err != nil {
			fail("write predictions: %v", err)
		}
	}
	if err := writeJSON(*outPath, report); err != nil {
		fail("write report: %v", err)
	}
	fmt.Printf("daily ranker leaves parity %s rows=%d max_abs_error=%.12g failures=%d report=%s\n",
		report.Status, len(predictions), maxAbsError, failures, *outPath)
	if report.Status != "passed" {
		os.Exit(1)
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

func featureColumnIndexes(header []string, features []string) ([]int, error) {
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

func columnIndex(header []string, name string) int {
	for i, column := range header {
		if column == name {
			return i
		}
	}
	return -1
}

func parseFeatureRow(row []string, indexes []int) ([]float64, error) {
	values := make([]float64, len(indexes))
	for i, idx := range indexes {
		if idx >= len(row) {
			return nil, fmt.Errorf("feature column %d outside row length %d", idx, len(row))
		}
		value, err := strconv.ParseFloat(row[idx], 64)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

func writePredictions(path string, rows []predictionRow) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"row", "expected_score", "go_score", "abs_error"}); err != nil {
		return err
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].AbsError > rows[j].AbsError
	})
	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.Itoa(row.Row),
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
