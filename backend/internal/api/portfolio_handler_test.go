package api

import "testing"

func TestLivePriceFromAlpacaQuoteUsesMidPrice(t *testing.T) {
	price, ok := livePriceFromAlpacaMessage("q", map[string]interface{}{
		"bp": 100.0,
		"ap": 102.0,
	})
	if !ok {
		t.Fatal("expected quote to produce a price")
	}
	if price != 101.0 {
		t.Fatalf("expected midpoint 101.0, got %f", price)
	}
}

func TestLivePriceFromAlpacaTradeUsesTradePrice(t *testing.T) {
	price, ok := livePriceFromAlpacaMessage("t", map[string]interface{}{
		"p": 99.25,
	})
	if !ok {
		t.Fatal("expected trade to produce a price")
	}
	if price != 99.25 {
		t.Fatalf("expected trade price 99.25, got %f", price)
	}
}
