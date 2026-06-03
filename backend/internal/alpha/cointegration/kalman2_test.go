package cointegration

import (
	"math"
	"testing"
)

func TestKalmanHedgeFilterConvergesOnSyntheticBeta(t *testing.T) {
	cfg := DefaultKalmanPairConfig("Y", "X")
	cfg.QAlpha = 1e-6
	cfg.QBeta = 1e-6
	cfg.R = 1e-4
	filter := NewKalmanHedgeFilter(cfg)

	trueAlpha := 0.4
	trueBeta := 1.35
	for i := 0; i < 250; i++ {
		x := 1.0 + float64(i)*0.02
		y := trueAlpha + trueBeta*x + math.Sin(float64(i))*0.0005
		if _, err := filter.Update(x, y); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	state := filter.State()
	if math.Abs(state.Beta-trueBeta) > 0.05 {
		t.Fatalf("beta=%f, want close to %f", state.Beta, trueBeta)
	}
}
