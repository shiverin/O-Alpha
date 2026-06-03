package cointegration

import (
	"context"
	"time"
)

const StrategyName = "kalman_cointegration"

type PairPositionState string

const (
	PairFlat        PairPositionState = "flat"
	PairShortYLongX PairPositionState = "short_y_long_x"
	PairLongYShortX PairPositionState = "long_y_short_x"
	PairQuarantined PairPositionState = "quarantined"
)

type KalmanPairConfig struct {
	SymbolY string
	SymbolX string

	QAlpha float64
	QBeta  float64
	R      float64

	EntryZ float64
	ExitZ  float64
	StopZ  float64

	MaxGrossWeight float64
	MaxLegWeight   float64

	RetestCadence              time.Duration
	ForceFlattenOnFailedRetest bool
	RequireShortable           bool
}

type KalmanPairState struct {
	Alpha                  float64
	Beta                   float64
	P00, P01, P10, P11     float64
	PositionState          PairPositionState
	LastZ                  float64
	LastInnovationVariance float64
	LastRetest             time.Time
	NeedsRetest            bool
	Quarantined            bool
	ProcessedBars          int
}

type ShortEligibilityProvider interface {
	IsShortable(ctx context.Context, symbol string, at time.Time) (bool, error)
}

type PairRetestProvider interface {
	PairStillApproved(ctx context.Context, symbolY, symbolX string, at time.Time) (bool, string, error)
}

func DefaultKalmanPairConfig(symbolY, symbolX string) KalmanPairConfig {
	return KalmanPairConfig{
		SymbolY:                    symbolY,
		SymbolX:                    symbolX,
		QAlpha:                     1e-5,
		QBeta:                      1e-5,
		R:                          1e-3,
		EntryZ:                     2.0,
		ExitZ:                      0.5,
		StopZ:                      3.75,
		MaxGrossWeight:             0.20,
		MaxLegWeight:               0.12,
		RetestCadence:              30 * 24 * time.Hour,
		ForceFlattenOnFailedRetest: true,
	}
}

func (c KalmanPairConfig) withDefaults() KalmanPairConfig {
	defaults := DefaultKalmanPairConfig(c.SymbolY, c.SymbolX)
	if c.QAlpha <= 0 {
		c.QAlpha = defaults.QAlpha
	}
	if c.QBeta <= 0 {
		c.QBeta = defaults.QBeta
	}
	if c.R <= 0 {
		c.R = defaults.R
	}
	if c.EntryZ <= 0 {
		c.EntryZ = defaults.EntryZ
	}
	if c.ExitZ < 0 {
		c.ExitZ = defaults.ExitZ
	}
	if c.StopZ <= c.EntryZ {
		c.StopZ = defaults.StopZ
	}
	if c.MaxGrossWeight <= 0 {
		c.MaxGrossWeight = defaults.MaxGrossWeight
	}
	if c.MaxGrossWeight > 1 {
		c.MaxGrossWeight = 1
	}
	if c.MaxLegWeight <= 0 {
		c.MaxLegWeight = defaults.MaxLegWeight
	}
	if c.MaxLegWeight > 1 {
		c.MaxLegWeight = 1
	}
	if c.RetestCadence <= 0 {
		c.RetestCadence = defaults.RetestCadence
	}
	return c
}
