package agent

import (
	"context"
	"testing"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestEnsembleStrictLookaheadBias(t *testing.T) {
	// Build a valid, non-zero sample history slice using a rolling pricing matrix
	bars := make([]models.Bar, 100)
	baseTime := time.Now().Truncate(time.Hour)
	for i := range bars {
		bars[i] = models.Bar{
			Time:   baseTime.Add(time.Duration(i) * time.Hour),
			Open:   100.0 + float64(i)*0.5,
			Close:  100.5 + float64(i)*0.5,
			High:   101.0 + float64(i)*0.5,
			Low:    99.9 + float64(i)*0.5,
			Volume: 50000,
		}
	}

	ensemble := NewEnsembleDecisionLayer(nil, nil, 50, RiskProfileModerate)
	ctx := context.Background()

	// Step 1: Evaluate state sequences locked to an early point in history (index 70)
	resAt70, err := ensemble.EvaluateLatest(ctx, bars[:71])
	assert.NoError(t, err)

	// Step 2: Inject subsequent updates to simulate forward progression up to index 90
	ensemble.Reset()
	var finalRes backtest.StrategyOutput
	for i := 50; i <= 90; i++ {
		finalRes, err = ensemble.EvaluateLatest(ctx, bars[:i+1])
		assert.NoError(t, err)

		if i == 70 {
			// Verify that forward progression matches point-in-time calculation exactly
			assert.Equal(t, resAt70.Signal, finalRes.Signal, "Lookahead Bias Detected: Past signals cannot change based on future parameters!")
			assert.Equal(t, resAt70.RegimeLabel, finalRes.RegimeLabel)
		}
	}
}
