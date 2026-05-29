package agent

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/oalpha/internal/alpaca"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
)

// AgentManager orchestrates the runtime lifecycles of live AgentWorkers.
type AgentManager struct {
	mu            sync.RWMutex
	activeAgents  map[string]*AgentWorker
	alpacaClient  *alpaca.Client
	repo          *db.BarsRepository
	portfolioRepo *db.PortfolioRepository
}

// NewAgentManager constructs an agent orchestrator.
func NewAgentManager(client *alpaca.Client, repo *db.BarsRepository, portfolioRepo *db.PortfolioRepository) *AgentManager {
	return &AgentManager{
		activeAgents:  make(map[string]*AgentWorker),
		alpacaClient:  client,
		repo:          repo,
		portfolioRepo: portfolioRepo,
	}
}

// GenerateKey builds a stable per-user, per-symbol worker key.
func (m *AgentManager) GenerateKey(userID int64, symbol string) string {
	return fmt.Sprintf("%d_%s", userID, symbol)
}

// StartAgent provisions and starts one background worker.
func (m *AgentManager) StartAgent(
	ctx context.Context,
	userID int64,
	symbol string,
	timeframe string,
	strat backtest.Strategy,
	paperTrade bool,
	initialCash float64,
	useWebSocket bool,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.GenerateKey(userID, symbol)
	if _, exists := m.activeAgents[key]; exists {
		return fmt.Errorf("agent for symbol %s is already running for this user", symbol)
	}

	worker := NewAgentWorker(
		context.Background(), // Decouple worker lifetime from the HTTP request context.
		m.alpacaClient,
		m.repo,
		m.portfolioRepo,
		userID,
		strat,
		symbol,
		timeframe,
		paperTrade,
		initialCash,
		useWebSocket,
	)

	if err := worker.Start(); err != nil {
		return fmt.Errorf("failed to start live agent: %w", err)
	}

	m.activeAgents[key] = worker

	go func(agentKey string, w *AgentWorker) {
		for {
			select {
			case <-w.Done():
				m.mu.Lock()
				delete(m.activeAgents, agentKey)
				m.mu.Unlock()
				return
			case err, ok := <-w.Err():
				if !ok {
					m.mu.Lock()
					delete(m.activeAgents, agentKey)
					m.mu.Unlock()
					return
				}
				if err != nil {
					log.Printf("Background agent worker error [%s]: %v", agentKey, err)
				}
			}
		}
	}(key, worker)

	return nil
}

// StopAgent terminates an active worker.
func (m *AgentManager) StopAgent(userID int64, symbol string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.GenerateKey(userID, symbol)
	worker, exists := m.activeAgents[key]
	if !exists {
		return fmt.Errorf("no active agent found running for symbol: %s", symbol)
	}

	worker.Stop()
	delete(m.activeAgents, key)
	return nil
}

// IsAgentRunning reports whether a worker is active.
func (m *AgentManager) IsAgentRunning(userID int64, symbol string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := m.GenerateKey(userID, symbol)
	_, exists := m.activeAgents[key]
	return exists
}
