package ml

import (
	"fmt"
	"math"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

const (
	BarrierProfitTake = "profit_take"
	BarrierStopLoss   = "stop_loss"
	BarrierVertical   = "vertical"
)

type PrimaryEvent struct {
	Symbol      string    `json:"symbol"`
	Time        time.Time `json:"time"`
	Index       int       `json:"index"`
	Side        int       `json:"side"`
	SignalScore float64   `json:"signal_score"`
	Price       float64   `json:"price"`
	Volatility  float64   `json:"volatility"`
}

type TripleBarrierConfig struct {
	HorizonBars          int     `json:"horizon_bars"`
	ProfitTakeVolMult    float64 `json:"profit_take_vol_mult"`
	StopLossVolMult      float64 `json:"stop_loss_vol_mult"`
	VolLookback          int     `json:"vol_lookback"`
	MinEventSpacingBars  int     `json:"min_event_spacing_bars"`
	EmbargoPct           float64 `json:"embargo_pct"`
	UseNextOpenForEntry  bool    `json:"use_next_open_for_entry"`
	VerticalBarrierLabel int     `json:"vertical_barrier_label"`
}

type MetaLabel struct {
	Event        PrimaryEvent `json:"event"`
	Label        int          `json:"label"`
	StartTime    time.Time    `json:"start_time"`
	EndTime      time.Time    `json:"end_time"`
	Barrier      string       `json:"barrier"`
	Return       float64      `json:"return"`
	SampleWeight float64      `json:"sample_weight"`
}

type LabelWindow struct {
	EventTime time.Time `json:"event_time"`
	EndTime   time.Time `json:"end_time"`
}

func DefaultTripleBarrierConfig() TripleBarrierConfig {
	return TripleBarrierConfig{
		HorizonBars:          5,
		ProfitTakeVolMult:    1.5,
		StopLossVolMult:      1.0,
		VolLookback:          20,
		MinEventSpacingBars:  1,
		EmbargoPct:           0.01,
		VerticalBarrierLabel: 0,
	}
}

func GeneratePrimaryEvents(bars []models.Bar, outputs []backtest.StrategyOutput, cfg TripleBarrierConfig) ([]PrimaryEvent, error) {
	if len(bars) == 0 {
		return nil, fmt.Errorf("cannot generate primary events from empty bars")
	}
	if len(outputs) != len(bars) {
		return nil, fmt.Errorf("strategy outputs length %d does not match bars length %d", len(outputs), len(bars))
	}
	cfg = cfg.withDefaults()

	events := make([]PrimaryEvent, 0)
	lastEventIndex := -cfg.MinEventSpacingBars - 1
	for i, output := range outputs {
		if output.Signal != models.SignalBuy {
			continue
		}
		if i-lastEventIndex < cfg.MinEventSpacingBars {
			continue
		}
		if bars[i].Close <= 0 {
			continue
		}
		events = append(events, PrimaryEvent{
			Symbol:      bars[i].Symbol,
			Time:        bars[i].Time,
			Index:       i,
			Side:        1,
			SignalScore: output.AlphaScore,
			Price:       bars[i].Close,
			Volatility:  LaggedRealizedVolatility(bars, i, cfg.VolLookback),
		})
		lastEventIndex = i
	}
	return events, nil
}

func LabelEvents(bars []models.Bar, events []PrimaryEvent, cfg TripleBarrierConfig) ([]MetaLabel, error) {
	if len(bars) == 0 {
		return nil, fmt.Errorf("cannot label events from empty bars")
	}
	cfg = cfg.withDefaults()

	labels := make([]MetaLabel, 0, len(events))
	for _, event := range events {
		index := event.Index
		if index < 0 {
			index = indexByTime(bars, event.Time)
		}
		if index < 0 || index >= len(bars) {
			return nil, fmt.Errorf("event at %s is outside bar series", event.Time)
		}
		label := labelOneEvent(bars, event, index, cfg)
		labels = append(labels, label)
	}
	return labels, nil
}

func LabelWindows(labels []MetaLabel) []LabelWindow {
	windows := make([]LabelWindow, len(labels))
	for i, label := range labels {
		windows[i] = LabelWindow{
			EventTime: label.StartTime,
			EndTime:   label.EndTime,
		}
	}
	return windows
}

func LaggedRealizedVolatility(bars []models.Bar, index, lookback int) float64 {
	if lookback <= 0 {
		lookback = 20
	}
	return closeToCloseVol(bars, index, lookback)
}

func labelOneEvent(bars []models.Bar, event PrimaryEvent, index int, cfg TripleBarrierConfig) MetaLabel {
	entryIndex := index
	entryPrice := event.Price
	if entryPrice <= 0 {
		entryPrice = bars[index].Close
	}
	if cfg.UseNextOpenForEntry && index+1 < len(bars) && bars[index+1].Open > 0 {
		entryIndex = index + 1
		entryPrice = bars[index+1].Open
	}
	startTime := bars[entryIndex].Time

	side := event.Side
	if side == 0 {
		side = 1
	}
	volatility := event.Volatility
	if volatility <= 0 {
		volatility = LaggedRealizedVolatility(bars, index, cfg.VolLookback)
	}
	profitBarrier := cfg.ProfitTakeVolMult * volatility
	stopBarrier := -cfg.StopLossVolMult * volatility
	horizonIndex := minInt(len(bars)-1, index+cfg.HorizonBars)

	endIndex := horizonIndex
	barrier := BarrierVertical
	pathReturn := pathAdjustedReturn(side, entryPrice, bars[horizonIndex].Close)
	label := cfg.VerticalBarrierLabel

	for i := entryIndex + 1; i <= horizonIndex; i++ {
		pathReturn = pathAdjustedReturn(side, entryPrice, bars[i].Close)
		if profitBarrier > 0 && pathReturn >= profitBarrier {
			endIndex = i
			barrier = BarrierProfitTake
			label = 1
			break
		}
		if stopBarrier < 0 && pathReturn <= stopBarrier {
			endIndex = i
			barrier = BarrierStopLoss
			label = 0
			break
		}
	}

	return MetaLabel{
		Event:        event,
		Label:        label,
		StartTime:    startTime,
		EndTime:      bars[endIndex].Time,
		Barrier:      barrier,
		Return:       pathReturn,
		SampleWeight: math.Abs(pathReturn),
	}
}

func pathAdjustedReturn(side int, entryPrice, exitPrice float64) float64 {
	if entryPrice <= 0 || exitPrice <= 0 {
		return 0
	}
	return float64(side) * (exitPrice/entryPrice - 1)
}

func indexByTime(bars []models.Bar, t time.Time) int {
	for i := range bars {
		if bars[i].Time.Equal(t) {
			return i
		}
	}
	return -1
}

func (c TripleBarrierConfig) withDefaults() TripleBarrierConfig {
	defaults := DefaultTripleBarrierConfig()
	if c.HorizonBars <= 0 {
		c.HorizonBars = defaults.HorizonBars
	}
	if c.ProfitTakeVolMult <= 0 {
		c.ProfitTakeVolMult = defaults.ProfitTakeVolMult
	}
	if c.StopLossVolMult <= 0 {
		c.StopLossVolMult = defaults.StopLossVolMult
	}
	if c.VolLookback <= 0 {
		c.VolLookback = defaults.VolLookback
	}
	if c.MinEventSpacingBars <= 0 {
		c.MinEventSpacingBars = defaults.MinEventSpacingBars
	}
	if c.EmbargoPct < 0 {
		c.EmbargoPct = 0
	}
	if c.EmbargoPct > 1 {
		c.EmbargoPct = 1
	}
	return c
}
