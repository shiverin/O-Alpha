package account

import (
	"context"
	"fmt"
	"sync"
)

// PaperAccount simulates a trading account for paper trading.
type PaperAccount struct {
	mu        sync.RWMutex
	Cash      float64
	Positions map[string]float64 // symbol -> number of shares
}

// NewPaperAccount creates a new paper trading account with initial cash.
func NewPaperAccount(initialCash float64) *PaperAccount {
	return &PaperAccount{
		Cash:      initialCash,
		Positions: make(map[string]float64),
	}
}

// Buy buys shares of a symbol at the given price.
// Returns the number of shares bought and the cost, or an error if insufficient funds.
func (a *PaperAccount) Buy(ctx context.Context, symbol string, price float64, amount float64) (float64, float64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if price <= 0 {
		return 0, 0, fmt.Errorf("price must be positive")
	}
	if amount <= 0 {
		return 0, 0, fmt.Errorf("amount must be positive")
	}
	cost := price * amount
	if a.Cash < cost {
		return 0, 0, fmt.Errorf("insufficient funds: need %.2f, have %.2f", cost, a.Cash)
	}
	a.Positions[symbol] += amount
	a.Cash -= cost
	return amount, cost, nil
}

// Sell sells shares of a symbol at the given price.
// Returns the number of shares sold and the proceeds, or an error if insufficient shares.
func (a *PaperAccount) Sell(ctx context.Context, symbol string, price float64, amount float64) (float64, float64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if price <= 0 {
		return 0, 0, fmt.Errorf("price must be positive")
	}
	if amount <= 0 {
		return 0, 0, fmt.Errorf("amount must be positive")
	}
	if a.Positions[symbol] < amount {
		return 0, 0, fmt.Errorf("insufficient position: have %.2f, need %.2f", a.Positions[symbol], amount)
	}
	a.Positions[symbol] -= amount
	if a.Positions[symbol] == 0 {
		delete(a.Positions, symbol)
	}
	proceeds := price * amount
	a.Cash += proceeds
	return amount, proceeds, nil
}

// Equity returns the total equity (cash + market value of positions) based on current prices.
func (a *PaperAccount) Equity(ctx context.Context, prices map[string]float64) float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	equity := a.Cash
	for symbol, shares := range a.Positions {
		if price, ok := prices[symbol]; ok {
			equity += price * shares
		}
	}
	return equity
}

// GetPosition returns the current position quantity for a symbol.
func (a *PaperAccount) GetPosition(symbol string) float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Positions[symbol]
}

// Snapshot returns a point-in-time copy of cash and positions.
func (a *PaperAccount) Snapshot() (float64, map[string]float64) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	positions := make(map[string]float64, len(a.Positions))
	for symbol, qty := range a.Positions {
		positions[symbol] = qty
	}

	return a.Cash, positions
}

func (a *PaperAccount) AvailableCash() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Cash
}
