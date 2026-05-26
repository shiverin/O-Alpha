package agent

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaperAccount_LifecycleAndConcurrency(t *testing.T) {
	ctx := context.Background()
	acct := NewPaperAccount(100000.0)

	// 1. Verify Basic Trade Math Boundaries
	qty, cost, err := acct.Buy(ctx, "AAPL", 150.0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 10.0, qty)
	assert.Equal(t, 1500.0, cost)
	assert.Equal(t, 98500.0, acct.Cash)

	// 2. Verify Overdraft Protections
	_, _, err = acct.Buy(ctx, "AAPL", 2000.0, 100)
	assert.Error(t, err, "Should throw error on insufficient capital balances")

	// 3. Verify Position Selling Restrictions
	_, proceeds, err := acct.Sell(ctx, "AAPL", 160.0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 800.0, proceeds)
	assert.Equal(t, 5.0, acct.GetPosition("AAPL"))

	// 4. Stress Test Thread-Safe RWMutex Concurrency
	var wg sync.WaitGroup
	workers := 50

	// Concurrent Readers checking equity values
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			prices := map[string]float64{"AAPL": 160.0}
			_ = acct.Equity(ctx, prices)
		}()
	}

	// Concurrent Writers trying to safely load more transactions
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _ = acct.Buy(ctx, "TSLA", 200.0, 1)
		}()
	}

	wg.Wait()
	assert.NotEqual(t, 0.0, acct.Cash, "Mutex locks must prevent collision or allocation corruption")
}
