package backtest

import "math"

type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

type FillPrice struct {
	Price          float64 `json:"price"`
	Notional       float64 `json:"notional"`
	SpreadCost     float64 `json:"spread_cost"`
	SlippageCost   float64 `json:"slippage_cost"`
	Commission     float64 `json:"commission"`
	SECFees        float64 `json:"sec_fees"`
	MarketImpact   float64 `json:"market_impact"`
	TotalCost      float64 `json:"total_cost"`
	AppliedCostBps float64 `json:"applied_cost_bps"`
}

type CostModel struct {
	SpreadBpsBySymbol       map[string]float64
	DefaultSpreadBps        float64
	SlippageBps             float64
	CommissionPerShare      float64
	MinCommission           float64
	BorrowFeeBpsAnnual      float64
	SECFeesBpsSell          float64
	EnableMarketImpact      bool
	MarketImpactCoefficient float64
}

func DefaultCostModel() CostModel {
	return CostModel{
		DefaultSpreadBps:   2,
		SlippageBps:        1,
		SECFeesBpsSell:     0,
		BorrowFeeBpsAnnual: 0,
	}
}

func (c CostModel) ApplyFill(side OrderSide, midOrOpen float64, notional float64, symbol string) FillPrice {
	if midOrOpen <= 0 || notional <= 0 {
		return FillPrice{}
	}
	spreadBps := c.DefaultSpreadBps
	if c.SpreadBpsBySymbol != nil {
		if override, ok := c.SpreadBpsBySymbol[symbol]; ok {
			spreadBps = override
		}
	}
	if spreadBps < 0 {
		spreadBps = 0
	}
	slippageBps := math.Max(0, c.SlippageBps)
	halfSpreadBps := spreadBps / 2
	impactBps := 0.0
	if c.EnableMarketImpact && c.MarketImpactCoefficient > 0 {
		impactBps = c.MarketImpactCoefficient * math.Sqrt(notional)
	}

	totalDirectionalBps := halfSpreadBps + slippageBps + impactBps
	var price float64
	switch side {
	case OrderSideBuy:
		price = midOrOpen * (1 + totalDirectionalBps/10_000)
	case OrderSideSell:
		price = midOrOpen * (1 - totalDirectionalBps/10_000)
	default:
		return FillPrice{}
	}

	quantity := notional / midOrOpen
	commission := math.Max(0, c.CommissionPerShare) * quantity
	if commission > 0 && c.MinCommission > 0 {
		commission = math.Max(commission, c.MinCommission)
	}
	secFees := 0.0
	if side == OrderSideSell && c.SECFeesBpsSell > 0 {
		secFees = notional * c.SECFeesBpsSell / 10_000
	}

	spreadCost := notional * halfSpreadBps / 10_000
	slippageCost := notional * slippageBps / 10_000
	marketImpact := notional * impactBps / 10_000
	totalCost := spreadCost + slippageCost + marketImpact + commission + secFees
	appliedCostBps := 0.0
	if notional > 0 {
		appliedCostBps = totalCost / notional * 10_000
	}
	return FillPrice{
		Price:          price,
		Notional:       notional,
		SpreadCost:     spreadCost,
		SlippageCost:   slippageCost,
		Commission:     commission,
		SECFees:        secFees,
		MarketImpact:   marketImpact,
		TotalCost:      totalCost,
		AppliedCostBps: appliedCostBps,
	}
}

func (c CostModel) BorrowCost(shortMarketValue float64, barsHeld int, barsPerYear float64) float64 {
	if shortMarketValue <= 0 || barsHeld <= 0 || c.BorrowFeeBpsAnnual <= 0 || barsPerYear <= 0 {
		return 0
	}
	return shortMarketValue * (c.BorrowFeeBpsAnnual / 10_000) * (float64(barsHeld) / barsPerYear)
}
