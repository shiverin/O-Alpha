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
	activeAgents  map[string]agentWorkerRuntime
	alpacaClient  *alpaca.Client
	repo          *db.BarsRepository
	agentRepo     *db.AgentRepository
	portfolioRepo *db.PortfolioRepository
}

type agentWorkerRuntime interface {
	Stop()
	Done() <-chan struct{}
	Err() <-chan error
	GetLatestMetrics() map[string]interface{}
}

// NewAgentManager constructs an agent orchestrator.
func NewAgentManager(client *alpaca.Client, repo *db.BarsRepository, agentRepo *db.AgentRepository, portfolioRepo *db.PortfolioRepository) *AgentManager {
	return &AgentManager{
		activeAgents:  make(map[string]agentWorkerRuntime),
		alpacaClient:  client,
		repo:          repo,
		agentRepo:     agentRepo,
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
	agentRunID int64,
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
		m.agentRepo,
		m.portfolioRepo,
		userID,
		strat,
		symbol,
		timeframe,
		paperTrade,
		initialCash,
		agentRunID,
		useWebSocket,
	)

	if err := worker.Start(); err != nil {
		return fmt.Errorf("failed to start live agent: %w", err)
	}

	m.activeAgents[key] = worker

	go m.monitorWorker(key, worker, "agent worker")

	return nil
}

// StartAgentV2 provisions the HMM ensemble through the unified worker runtime.
func (m *AgentManager) StartAgentV2(
	ctx context.Context,
	userID int64,
	symbol string,
	timeframe string,
	maFastPeriod int,
	maSlowPeriod int,
	kalmanQNoise float64,
	kalmanRNoise float64,
	kalmanZThreshold float64,
	paperTrade bool,
	initialCash float64,
	agentRunID int64,
	riskProfile RiskProfile,
	useWebSocket bool,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.GenerateKey(userID, symbol)
	if _, exists := m.activeAgents[key]; exists {
		return fmt.Errorf("agent for symbol %s is already running for this user", symbol)
	}

	maStrat := backtest.NewMACrossoverStrategy(maFastPeriod, maSlowPeriod)
	kalmanStrat := backtest.NewKalmanStrategy(kalmanQNoise, kalmanRNoise, 20, kalmanZThreshold)
	ensemble := NewEnsembleDecisionLayer(maStrat, kalmanStrat, 50, riskProfile)
	worker := NewAgentWorker(
		context.Background(), // Decouple worker lifetime from the HTTP request context.
		m.alpacaClient,
		m.repo,
		m.agentRepo,
		m.portfolioRepo,
		userID,
		ensemble,
		symbol,
		timeframe,
		paperTrade,
		initialCash,
		agentRunID,
		useWebSocket,
	)

	if err := worker.Start(); err != nil {
		return fmt.Errorf("failed to start HMM ensemble agent: %w", err)
	}

	m.activeAgents[key] = worker
	go m.monitorWorker(key, worker, "HMM ensemble agent")

	return nil
}

func (m *AgentManager) monitorWorker(agentKey string, worker agentWorkerRuntime, label string) {
	for {
		select {
		case <-worker.Done():
			m.mu.Lock()
			delete(m.activeAgents, agentKey)
			m.mu.Unlock()
			return
		case err, ok := <-worker.Err():
			if !ok {
				m.mu.Lock()
				delete(m.activeAgents, agentKey)
				m.mu.Unlock()
				return
			}
			if err != nil {
				log.Printf("Background %s error [%s]: %v", label, agentKey, err)
			}
		}
	}
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

func (w *AgentWorker) GetLatestMetrics() map[string]interface{} {
	metricsMap := make(map[string]interface{})
	w.telemetryMetadata.Range(func(key, value interface{}) bool {
		if keyString, ok := key.(string); ok {
			metricsMap[keyString] = value
		}
		return true
	})
	return metricsMap
}
