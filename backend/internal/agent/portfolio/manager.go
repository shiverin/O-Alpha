package portfolio

import (
	"context"
	"fmt"
	"sync"

	"github.com/oalpha/internal/alpaca"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
)

type PortfolioAgentManager struct {
	mu            sync.RWMutex
	activeWorkers map[string]*PortfolioAgentWorker
	repo          *db.BarsRepository
	alpacaClient  *alpaca.Client
}

func NewPortfolioAgentManager(repo *db.BarsRepository, alpacaClient *alpaca.Client) *PortfolioAgentManager {
	return &PortfolioAgentManager{
		activeWorkers: make(map[string]*PortfolioAgentWorker),
		repo:          repo,
		alpacaClient:  alpacaClient,
	}
}

func (m *PortfolioAgentManager) StartPortfolioAgent(
	ctx context.Context,
	key string,
	strategy backtest.PortfolioStrategy,
	symbols []string,
	timeframe string,
	initialCash float64,
	execution ExecutionRouter,
) (*PortfolioAgentWorker, error) {
	if key == "" {
		return nil, fmt.Errorf("portfolio agent key is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.activeWorkers[key]; exists {
		return nil, fmt.Errorf("portfolio agent %s is already running", key)
	}
	worker := NewPortfolioAgentWorker(ctx, strategy, symbols, timeframe, initialCash, m.repo, m.alpacaClient, execution)
	m.activeWorkers[key] = worker
	return worker, nil
}

func (m *PortfolioAgentManager) StopPortfolioAgent(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	worker, exists := m.activeWorkers[key]
	if !exists {
		return fmt.Errorf("portfolio agent %s is not running", key)
	}
	worker.Stop()
	delete(m.activeWorkers, key)
	return nil
}
