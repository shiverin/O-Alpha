package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/oalpha/internal/ml"
)

func main() {
	metadataPath := flag.String("metadata", "", "path to model metadata.json")
	fixturePath := flag.String("fixture", "", "path to parity_fixture.csv")
	outPath := flag.String("out", "", "optional output CSV path for Go predictions")
	tolerance := flag.Float64("tolerance", 1e-6, "probability parity tolerance")
	flag.Parse()

	if *metadataPath == "" || *fixturePath == "" {
		fail("usage: validate-leaves-parity --metadata metadata.json --fixture parity_fixture.csv [--out go_predictions.csv]")
	}

	artifact, err := ml.ReadModelArtifact(*metadataPath)
	if err != nil {
		fail("read metadata: %v", err)
	}
	root := filepath.Dir(*metadataPath)
	predictor, err := ml.NewLeavesPredictor(artifact.ModelPath(root), artifact.FeatureSpec, artifact.Version())
	if err != nil {
		fail("load predictor: %v", err)
	}

	rows, header, err := readCSV(*fixturePath)
	if err != nil {
		fail("read fixture: %v", err)
	}
	featureIndexes, err := featureColumnIndexes(header, artifact.FeatureSpec.Features)
	if err != nil {
		fail("validate fixture columns: %v", err)
	}
	expectedIndex := columnIndex(header, "expected_probability")

	goProbabilities := make([]float64, len(rows))
	expected := make([]float64, 0, len(rows))
	actualForParity := make([]float64, 0, len(rows))
	for i, row := range rows {
		features, err := parseFeatureRow(row, featureIndexes)
		if err != nil {
			fail("parse row %d: %v", i, err)
		}
		probability, err := predictor.PredictProba(features)
		if err != nil {
			fail("predict row %d: %v", i, err)
		}
		goProbabilities[i] = probability
		if expectedIndex >= 0 {
			p, err := strconv.ParseFloat(row[expectedIndex], 64)
			if err != nil {
				fail("parse expected probability row %d: %v", i, err)
			}
			expected = append(expected, p)
			actualForParity = append(actualForParity, probability)
		}
	}

	if *outPath != "" {
		if err := writePredictions(*outPath, goProbabilities); err != nil {
			fail("write predictions: %v", err)
		}
	}
	if len(expected) > 0 {
		maxAbsError, err := ml.ValidateParity(expected, actualForParity, *tolerance)
		if err != nil {
			fail("%v", err)
		}
		fmt.Printf("leaves parity passed max_abs_error=%.12g tolerance=%.12g\n", maxAbsError, *tolerance)
		return
	}
	fmt.Printf("wrote %d Go predictions\n", len(goProbabilities))
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
			return nil, fmt.Errorf("missing feature column %s", feature)
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

func writePredictions(path string, probabilities []float64) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"row", "go_probability"}); err != nil {
		return err
	}
	for i, probability := range probabilities {
		if err := writer.Write([]string{strconv.Itoa(i), strconv.FormatFloat(probability, 'g', -1, 64)}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
