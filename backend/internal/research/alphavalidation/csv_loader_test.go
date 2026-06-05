package alphavalidation

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadBarsCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bars.csv")
	content := "time,symbol,open,high,low,close,volume\n2024-01-01,VOO,100,101,99,100,1000\n2024-01-01,AAPL,50,51,49,50,1000\n2024-01-02,VOO,101,102,100,101,1000\n2024-01-02,AAPL,51,52,50,51,1000\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	window := ValidationWindow{
		From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 23, 59, 59, 0, time.UTC),
	}
	bars, err := LoadBarsCSV(path, []string{"VOO", "AAPL"}, window)
	if err != nil {
		t.Fatalf("load csv: %v", err)
	}
	if len(bars["VOO"]) != 2 || len(bars["AAPL"]) != 2 {
		t.Fatalf("unexpected bars loaded: %+v", bars)
	}
}
