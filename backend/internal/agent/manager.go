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
	mu           sync.RWMutex
	activeAgents map[string]*AgentWorker
	alpacaClient *alpaca.Client
	repo         *db.Repository
}

// NewAgentManager instantiates a clean orchestrator instance.
func NewAgentManager(client *alpaca.Client, repo *db.Repository) *AgentManager {
	return &AgentManager{
		activeAgents: make(map[string]*AgentWorker),
		alpacaClient: client,
		repo:         repo,
	}
}

// GenerateKey builds a unique tracking handle for an agent loop.
func (m *AgentManager) GenerateKey(userID int64, symbol string) string {
	return fmt.Sprintf("%d_%s", userID, symbol)
}

// StartAgent provisions, warms up, and initializes a background market routing pipeline.
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

	// Spin up the worker thread using your existing constructor layout
	worker := NewAgentWorker(
		context.Background(), // Root context detached from transient HTTP lifecycle boundaries
		m.alpacaClient,
		m.repo,
		strat,
		symbol,
		timeframe,
		paperTrade,
		initialCash,
		useWebSocket,
	)

	// Trigger indicators warmup and start polling loops
	if err := worker.Start(); err != nil {
		return fmt.Errorf("failed to start live agent: %w", err)
	}

	m.activeAgents[key] = worker
	m.activeAgents[key] = worker

	// Background monitor loop to auto-cleanup dropped/failed worker nodes safely
	go func(agentKey string, w *AgentWorker) {
		select {
		case <-w.Done():
		case err := <-w.Err():
			if err != nil {
				// FIX: Added a valid log operation to eliminate the empty branch lint warning
				log.Printf("Background agent worker error [%s]: %v", agentKey, err)
			}
		}
		m.mu.Lock()
		delete(m.activeAgents, agentKey)
		m.mu.Unlock()
	}(key, worker)

	return nil
}

// StopAgent gracefully terminates an active streaming thread channel.
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

// IsAgentRunning handles quick real-time visual flag checks for dashboard widgets.
func (m *AgentManager) IsAgentRunning(userID int64, symbol string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := m.GenerateKey(userID, symbol)
	_, exists := m.activeAgents[key]
	return exists
}
