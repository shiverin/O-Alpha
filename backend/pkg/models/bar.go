package models

import "time"

// Bar is a single OHLCV candle.
type Bar struct {
	Time   time.Time `json:"time"`
	Symbol string    `json:"symbol"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume int64     `json:"volume"`
}

// Signal is a trading action.
type Signal int

const (
	SignalHold Signal = iota
	SignalBuy
	SignalSell
)

// BacktestRequest configures a MA crossover backtest run.
type BacktestRequest struct {
	Symbol      string     `json:"symbol" binding:"required"`
	FastPeriod  int        `json:"fast_period" binding:"required,min=1"`
	SlowPeriod  int        `json:"slow_period" binding:"required,min=2"`
	InitialCash float64    `json:"initial_cash"`
	Start       *time.Time `json:"start,omitempty"`
	End         *time.Time `json:"end,omitempty"`
}

// EquityPoint is one point on the equity curve.
type EquityPoint struct {
	Time   time.Time `json:"time"`
	Equity float64   `json:"equity"`
}

// BacktestResult is the output of a backtest run.
type BacktestResult struct {
	Symbol      string        `json:"symbol"`
	EquityCurve []EquityPoint `json:"equity_curve"`
	FinalEquity float64       `json:"final_equity"`
	TotalReturn float64       `json:"total_return"`
	Sharpe      float64       `json:"sharpe"`
	Sortino     float64       `json:"sortino"`
	MaxDrawdown float64       `json:"max_drawdown"`
	NumTrades   int           `json:"num_trades"`
}
