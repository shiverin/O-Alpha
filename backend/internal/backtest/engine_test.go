package backtest

import (
	"testing"
	"time"

	"github.com/oalpha/pkg/models"
)

func TestRunBacktestWithOutputsPreservesIdleCash(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := []models.Bar{
		{Time: start, Symbol: "TEST", Open: 100, High: 100, Low: 100, Close: 100},
		{Time: start.Add(time.Hour), Symbol: "TEST", Open: 100, High: 100, Low: 100, Close: 100},
		{Time: start.Add(2 * time.Hour), Symbol: "TEST", Open: 110, High: 110, Low: 110, Close: 110},
		{Time: start.Add(3 * time.Hour), Symbol: "TEST", Open: 120, High: 120, Low: 120, Close: 120},
	}
	outputs := []StrategyOutput{
		{Signal: models.SignalBuy, PositionSizePct: 0.10},
		{Signal: models.SignalHold},
		{Signal: models.SignalSell},
		{Signal: models.SignalHold},
	}

	result, err := RunBacktestWithOutputs(bars, outputs, 1000)
	if err != nil {
		t.Fatalf("backtest: %v", err)
	}
	if result.FinalEquity != 1020 {
		t.Fatalf("expected idle cash to be preserved, final equity = %.2f", result.FinalEquity)
	}
	if result.ProfitFactor != 0 {
		t.Fatalf("profit factor should be finite when there are no losses, got %.2f", result.ProfitFactor)
	}
	if result.WinRate != 1 {
		t.Fatalf("expected one winning closed trade, win rate = %.2f", result.WinRate)
	}
	if len(result.Trades) != 1 {
		t.Fatalf("expected one closed trade, got %d", len(result.Trades))
	}
	trade := result.Trades[0]
	if trade.EntryPrice != 100 || trade.ExitPrice != 120 {
		t.Fatalf("unexpected trade prices: entry %.2f exit %.2f", trade.EntryPrice, trade.ExitPrice)
	}
}

func TestRunBuyAndHoldUsesFullAllocation(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := []models.Bar{
		{Time: start, Symbol: "TEST", Open: 100, High: 100, Low: 100, Close: 100},
		{Time: start.Add(time.Hour), Symbol: "TEST", Open: 100, High: 100, Low: 100, Close: 100},
		{Time: start.Add(2 * time.Hour), Symbol: "TEST", Open: 110, High: 110, Low: 110, Close: 110},
	}

	result, err := RunBuyAndHold(bars, 1000)
	if err != nil {
		t.Fatalf("buy and hold: %v", err)
	}
	if result.FinalEquity != 1100 {
		t.Fatalf("expected buy-and-hold final equity 1100, got %.2f", result.FinalEquity)
	}
	if result.NumTrades != 2 {
		t.Fatalf("expected buy and final liquidation trades, got %d", result.NumTrades)
	}
	if len(result.Trades) != 1 {
		t.Fatalf("expected one closed buy-and-hold trade, got %d", len(result.Trades))
	}
	if result.Trades[0].EntryPrice != 100 || result.Trades[0].ExitPrice != 110 {
		t.Fatalf("unexpected buy-and-hold trade prices: entry %.2f exit %.2f", result.Trades[0].EntryPrice, result.Trades[0].ExitPrice)
	}
}
