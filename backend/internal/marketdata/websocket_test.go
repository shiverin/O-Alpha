package marketdata

import "testing"

func TestParseAlpacaBarSkipsMalformedMessages(t *testing.T) {
	connector := &WebSocketConnector{}

	cases := []map[string]interface{}{
		{"T": "success", "msg": "authenticated"},
		{"T": "b", "S": "AAPL"},
		{"T": "b", "S": "AAPL", "t": "bad-time", "o": 1.0, "h": 2.0, "l": 1.0, "c": 1.5, "v": 100.0},
	}

	for _, tc := range cases {
		if bar := connector.parseAlpacaBar(tc); bar != nil {
			t.Fatalf("expected malformed message to be skipped, got %#v", bar)
		}
	}
}

func TestParseAlpacaBarParsesValidBar(t *testing.T) {
	connector := &WebSocketConnector{}

	bar := connector.parseAlpacaBar(map[string]interface{}{
		"T": "b",
		"S": "aapl",
		"t": "2026-05-29T10:00:00Z",
		"o": 100.0,
		"h": 105.0,
		"l": 99.0,
		"c": 103.5,
		"v": 1000.0,
	})
	if bar == nil {
		t.Fatal("expected valid bar to parse")
	}
	if bar.Symbol != "AAPL" {
		t.Fatalf("expected uppercase symbol, got %s", bar.Symbol)
	}
	if bar.Close != 103.5 {
		t.Fatalf("expected close 103.5, got %f", bar.Close)
	}
}
