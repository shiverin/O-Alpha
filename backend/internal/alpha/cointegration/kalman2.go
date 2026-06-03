package cointegration

import (
	"fmt"
	"math"
)

type KalmanHedgeFilter struct {
	state       KalmanPairState
	qAlpha      float64
	qBeta       float64
	r           float64
	initialized bool
}

type KalmanUpdate struct {
	Alpha              float64
	Beta               float64
	Innovation         float64
	InnovationVariance float64
	Z                  float64
	Spread             float64
}

func NewKalmanHedgeFilter(cfg KalmanPairConfig) *KalmanHedgeFilter {
	cfg = cfg.withDefaults()
	return &KalmanHedgeFilter{
		state: KalmanPairState{
			Beta:          1,
			P00:           1,
			P11:           1,
			PositionState: PairFlat,
		},
		qAlpha: cfg.QAlpha,
		qBeta:  cfg.QBeta,
		r:      cfg.R,
	}
}

func (f *KalmanHedgeFilter) Update(logX, logY float64) (KalmanUpdate, error) {
	if f == nil {
		return KalmanUpdate{}, fmt.Errorf("kalman hedge filter is nil")
	}
	if !isFinitePositive(logX) || !isFinitePositive(logY) {
		return KalmanUpdate{}, fmt.Errorf("log prices must be finite")
	}
	if !f.initialized {
		f.state.Alpha = logY - f.state.Beta*logX
		f.initialized = true
	}

	p00 := f.state.P00 + f.qAlpha
	p01 := f.state.P01
	p10 := f.state.P10
	p11 := f.state.P11 + f.qBeta

	predictedY := f.state.Alpha + f.state.Beta*logX
	innovation := logY - predictedY
	innovationVariance := p00 + p01*logX + logX*p10 + logX*logX*p11 + f.r
	if innovationVariance <= 0 || math.IsNaN(innovationVariance) || math.IsInf(innovationVariance, 0) {
		return KalmanUpdate{}, fmt.Errorf("invalid innovation variance %f", innovationVariance)
	}

	k0 := (p00 + p01*logX) / innovationVariance
	k1 := (p10 + p11*logX) / innovationVariance

	alpha := f.state.Alpha + k0*innovation
	beta := f.state.Beta + k1*innovation

	// P = (I - K H) P, H = [1, logX].
	newP00 := (1-k0)*p00 - k0*logX*p10
	newP01 := (1-k0)*p01 - k0*logX*p11
	newP10 := -k1*p00 + (1-k1*logX)*p10
	newP11 := -k1*p01 + (1-k1*logX)*p11

	f.state.Alpha = alpha
	f.state.Beta = beta
	f.state.P00 = newP00
	f.state.P01 = newP01
	f.state.P10 = newP10
	f.state.P11 = newP11
	f.state.LastInnovationVariance = innovationVariance
	f.state.LastZ = innovation / math.Sqrt(innovationVariance)

	return KalmanUpdate{
		Alpha:              alpha,
		Beta:               beta,
		Innovation:         innovation,
		InnovationVariance: innovationVariance,
		Z:                  f.state.LastZ,
		Spread:             logY - alpha - beta*logX,
	}, nil
}

func (f *KalmanHedgeFilter) State() KalmanPairState {
	if f == nil {
		return KalmanPairState{}
	}
	return f.state
}

func (f *KalmanHedgeFilter) SetState(state KalmanPairState) {
	if f == nil {
		return
	}
	f.state = state
	f.initialized = true
}

func isFinitePositive(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
