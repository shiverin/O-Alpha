package backtest

import "testing"

func TestCostModelApplyFillChargesSpreadAndSlippage(t *testing.T) {
	model := CostModel{
		DefaultSpreadBps: 4,
		SlippageBps:      1,
	}

	fill := model.ApplyFill(OrderSideBuy, 100, 10_000, "SPY")
	if fill.Price != 100.03 {
		t.Fatalf("expected buy fill price 100.03, got %.4f", fill.Price)
	}
	if fill.SpreadCost != 2 {
		t.Fatalf("expected half-spread cost 2.00, got %.4f", fill.SpreadCost)
	}
	if fill.SlippageCost != 1 {
		t.Fatalf("expected slippage cost 1.00, got %.4f", fill.SlippageCost)
	}
}
