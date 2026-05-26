package agent

import (
	"testing"
	"time"

	"github.com/oalpha/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAgentWorker_HistoricalBufferTrimming(t *testing.T) {
	// Provision a dummy worker instance targeting short max bounds
	worker := &AgentWorker{
		maxBars:        5,
		historicalBars: make([]models.Bar, 0),
	}

	now := time.Now()
	// Hydrate 7 sequential candles into the engine map manually
	for i := 0; i < 7; i++ {
		worker.appendOrUpdateBar(models.Bar{
			Time:  now.Add(time.Duration(i) * time.Minute),
			Close: 100.0 + float64(i),
		})
	}

	// Verify that data management bounds correctly drop historical bloat
	bars := worker.getHistoricalBarsSnapshot()
	assert.Equal(t, 5, len(bars), "Worker buffer must cleanly truncate oldest elements to prevent memory leakage loops")
	assert.Equal(t, 106.0, bars[4].Close, "Latest tick data properties must remain uncorrupted at the tail index")
}
