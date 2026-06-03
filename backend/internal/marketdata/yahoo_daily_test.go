package marketdata

import (
	"encoding/json"
	"testing"
)

func TestParseYahooDailyBarsAdjustsOHLC(t *testing.T) {
	payload := yahooChartResponse{}
	raw := []byte(`{
		"chart": {
			"result": [{
				"timestamp": [1420209000, 1420468200],
				"indicators": {
					"quote": [{
						"open": [100.0, null],
						"high": [110.0, 120.0],
						"low": [90.0, 100.0],
						"close": [105.0, 110.0],
						"volume": [1000.0, 2000.0]
					}],
					"adjclose": [{
						"adjclose": [52.5, 55.0]
					}]
				}
			}],
			"error": null
		}
	}`)
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	bars, err := ParseYahooDailyBars("aapl", payload)
	if err != nil {
		t.Fatalf("parse bars: %v", err)
	}
	if len(bars) != 1 {
		t.Fatalf("expected one complete bar, got %d", len(bars))
	}
	if bars[0].Symbol != "AAPL" {
		t.Fatalf("expected uppercase symbol, got %s", bars[0].Symbol)
	}
	if bars[0].Open != 50 || bars[0].High != 55 || bars[0].Low != 45 || bars[0].Close != 52.5 {
		t.Fatalf("unexpected adjusted OHLC: %#v", bars[0])
	}
	if bars[0].Volume != 1000 {
		t.Fatalf("expected volume 1000, got %d", bars[0].Volume)
	}
}

func TestParseYahooDailyBarsReturnsChartError(t *testing.T) {
	payload := yahooChartResponse{}
	raw := []byte(`{"chart":{"result":null,"error":{"code":"Not Found","description":"No data found"}}}`)
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	if _, err := ParseYahooDailyBars("BAD", payload); err == nil {
		t.Fatal("expected chart error")
	}
}

func TestParseYahooDailyBarsClampsAdjustedHighLow(t *testing.T) {
	payload := yahooChartResponse{}
	raw := []byte(`{
		"chart": {
			"result": [{
				"timestamp": [1420209000],
				"indicators": {
					"quote": [{
						"open": [100.0],
						"high": [101.0],
						"low": [99.0],
						"close": [100.0],
						"volume": [1000.0]
					}],
					"adjclose": [{
						"adjclose": [105.0]
					}]
				}
			}],
			"error": null
		}
	}`)
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	bars, err := ParseYahooDailyBars("VOO", payload)
	if err != nil {
		t.Fatalf("parse bars: %v", err)
	}
	if len(bars) != 1 {
		t.Fatalf("expected one bar, got %d", len(bars))
	}
	if bars[0].High < bars[0].Close || bars[0].Low > bars[0].Open {
		t.Fatalf("expected OHLC-consistent bar, got %#v", bars[0])
	}
}
