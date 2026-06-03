package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReportBaseNameShortensLongSymbolLists(t *testing.T) {
	symbols := strings.Split("VOO,AAPL,MSFT,NVDA,AMZN,META,GOOGL,AVGO,LLY,JPM,V,UNH,HD,MA,COST,PG,WMT,NFLX,CRM,ORCL,AMD,PEP,KO,MRK,ABBV,ABT,ACN,ADBE,AMAT,AMGN,AXP,BA,BAC,BKNG,BLK,BMY,C,CAT,CI,CMCSA,COP,CSCO,CVS,CVX,DE,DIS,ELV,GE,GILD,GS,HON,IBM,INTC,INTU,ISRG,JNJ,LIN,LOW,LRCX,MCD,MDLZ,MDT,MO,NKE,NOW,PFE,PLD,PM,QCOM,RTX,SBUX,SCHW,SO,SYK,T,TMO,TXN,UPS,USB,VZ,XOM", ",")

	name := reportBaseName(symbols, "1Day")
	if len(name) > 200 {
		t.Fatalf("report base name length=%d, want <= 200: %s", len(name), name)
	}
	if !strings.Contains(name, fmt.Sprintf("%dsymbols", len(symbols))) {
		t.Fatalf("report base name should include symbol count, got %s", name)
	}
	if !strings.HasSuffix(name, "_1day_alpha_validation") {
		t.Fatalf("report base name suffix mismatch: %s", name)
	}
}

func TestLoadPanelFromCSVFiltersAndAlignsSymbols(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bars.csv")
	csv := strings.Join([]string{
		"symbol,time,open,high,low,close,volume",
		"VOO,2024-01-02T14:30:00Z,100,101,99,100.5,1000",
		"AAPL,2024-01-02T14:30:00Z,10,11,9,10.5,2000",
		"MSFT,2024-01-02T14:30:00Z,20,21,19,20.5,3000",
		"VOO,2024-01-03T14:30:00Z,101,102,100,101.5,1000",
		"AAPL,2024-01-03T14:30:00Z,11,12,10,11.5,2000",
		"MSFT,2024-01-03T14:30:00Z,21,22,20,21.5,3000",
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}

	start, _ := time.Parse("2006-01-02", "2024-01-01")
	end, _ := time.Parse("2006-01-02", "2024-01-04")
	panel, err := loadPanelFromCSV(path, []string{"VOO", "AAPL"}, "1Day", start, end)
	if err != nil {
		t.Fatalf("loadPanelFromCSV returned error: %v", err)
	}
	panel.Symbols = orderSymbols(panel.Symbols, []string{"VOO", "AAPL"})

	if got, want := strings.Join(panel.Symbols, ","), "VOO,AAPL"; got != want {
		t.Fatalf("symbols=%s, want %s", got, want)
	}
	if len(panel.Times) != 2 {
		t.Fatalf("times=%d, want 2", len(panel.Times))
	}
	if panel.Bars["VOO"][1].Close != 101.5 {
		t.Fatalf("VOO close=%v, want 101.5", panel.Bars["VOO"][1].Close)
	}
	if _, ok := panel.Bars["MSFT"]; ok {
		t.Fatalf("unexpected filtered symbol MSFT in panel")
	}
}
