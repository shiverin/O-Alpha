package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AgentSettings maps to a user's saved trading parameters.
type AgentSettings struct {
	UserID        int64     `json:"user_id"`
	RiskProfile   string    `json:"risk_profile"`
	Leverage      int       `json:"leverage"`
	MaxPositions  int       `json:"max_positions"`
	StopLossPct   float64   `json:"stop_loss_pct"`
	TakeProfitPct float64   `json:"take_profit_pct"`
	RebalanceFreq string    `json:"rebalance_freq"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AgentRepository persists user trading parameters.
type AgentRepository struct {
	db *pgxpool.Pool
}

// NewAgentRepository creates an agent settings repository.
func NewAgentRepository(db *pgxpool.Pool) *AgentRepository {
	return &AgentRepository{db: db}
}

// GetAgentSettings returns saved settings for a user.
func (r *AgentRepository) GetAgentSettings(ctx context.Context, userID int64) (*AgentSettings, error) {
	const q = `
		SELECT user_id, risk_profile, leverage, max_positions, stop_loss_pct, take_profit_pct, rebalance_freq, updated_at
		FROM agent_settings
		WHERE user_id = $1`

	var s AgentSettings
	err := r.db.QueryRow(ctx, q, userID).Scan(
		&s.UserID, &s.RiskProfile, &s.Leverage, &s.MaxPositions, &s.StopLossPct, &s.TakeProfitPct, &s.RebalanceFreq, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select agent settings: %w", err)
	}
	return &s, nil
}

// SaveAgentSettings creates or updates saved settings with an upsert.
func (r *AgentRepository) SaveAgentSettings(ctx context.Context, s *AgentSettings) error {
	const q = `
		INSERT INTO agent_settings (user_id, risk_profile, leverage, max_positions, stop_loss_pct, take_profit_pct, rebalance_freq, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			risk_profile = EXCLUDED.risk_profile,
			leverage = EXCLUDED.leverage,
			max_positions = EXCLUDED.max_positions,
			stop_loss_pct = EXCLUDED.stop_loss_pct,
			take_profit_pct = EXCLUDED.take_profit_pct,
			rebalance_freq = EXCLUDED.rebalance_freq,
			updated_at = NOW()`

	_, err := r.db.Exec(ctx, q, s.UserID, s.RiskProfile, s.Leverage, s.MaxPositions, s.StopLossPct, s.TakeProfitPct, s.RebalanceFreq)
	if err != nil {
		return fmt.Errorf("upsert agent settings: %w", err)
	}
	return nil
}

// CreateAgentRun provisions a persisted runtime record for a paper or live agent.
func (r *AgentRepository) CreateAgentRun(
	ctx context.Context,
	userID int64,
	symbol string,
	strategyType string,
	timeframe string,
	mode string,
	initialCash float64,
	useWebSocket bool,
	parameters map[string]interface{},
) (int64, error) {
	symbol = normalizeSymbol(symbol)
	strategyType = strings.ToUpper(strings.TrimSpace(strategyType))
	mode = strings.ToLower(strings.TrimSpace(mode))
	if symbol == "" {
		return 0, fmt.Errorf("agent run symbol is required")
	}
	if mode == "" {
		mode = "paper"
	}
	if initialCash <= 0 {
		initialCash = defaultPaperInitialCash
	}

	paramsBytes, err := json.Marshal(parameters)
	if err != nil {
		return 0, fmt.Errorf("marshal agent run parameters: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin agent run transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := ensureAssetTx(ctx, tx, symbol); err != nil {
		return 0, err
	}

	account, err := ensureDefaultPaperAccountTx(ctx, tx, userID, initialCash)
	if err != nil {
		return 0, err
	}

	var runID int64
	const q = `
		INSERT INTO agent_runs (
			user_id,
			account_id,
			symbol,
			strategy_type,
			timeframe,
			mode,
			status,
			initial_cash,
			use_websocket,
			parameters
		)
		VALUES ($1, $2, $3, $4, $5, $6, 'starting', $7, $8, $9)
		RETURNING id`
	if err := tx.QueryRow(ctx, q, userID, account.ID, symbol, strategyType, timeframe, mode, initialCash, useWebSocket, paramsBytes).Scan(&runID); err != nil {
		return 0, fmt.Errorf("insert agent run: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit agent run transaction: %w", err)
	}
	return runID, nil
}

// MarkAgentRunRunning records that the background worker started successfully.
func (r *AgentRepository) MarkAgentRunRunning(ctx context.Context, runID int64) error {
	const q = `
		UPDATE agent_runs
		SET status = 'running',
			last_heartbeat_at = NOW()
		WHERE id = $1 AND status = 'starting'`
	tag, err := r.db.Exec(ctx, q, runID)
	if err != nil {
		return fmt.Errorf("mark agent run running: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("agent run %d was not found in starting status", runID)
	}
	return nil
}

// MarkAgentRunFailed closes a run that could not start or continue.
func (r *AgentRepository) MarkAgentRunFailed(ctx context.Context, runID int64, reason string) error {
	const q = `
		UPDATE agent_runs
		SET status = 'failed',
			stopped_at = NOW(),
			stop_reason = $2
		WHERE id = $1 AND status IN ('starting', 'running', 'stopping')`
	_, err := r.db.Exec(ctx, q, runID, reason)
	if err != nil {
		return fmt.Errorf("mark agent run failed: %w", err)
	}
	return nil
}

// MarkLatestAgentRunStopped closes the most recent active run for a user and symbol.
func (r *AgentRepository) MarkLatestAgentRunStopped(ctx context.Context, userID int64, symbol string, reason string) error {
	symbol = normalizeSymbol(symbol)
	const q = `
		WITH latest AS (
			SELECT id
			FROM agent_runs
			WHERE user_id = $1
				AND symbol = $2
				AND status IN ('starting', 'running', 'stopping')
			ORDER BY started_at DESC
			LIMIT 1
		)
		UPDATE agent_runs
		SET status = 'stopped',
			stopped_at = NOW(),
			stop_reason = $3
		WHERE id = (SELECT id FROM latest)`
	_, err := r.db.Exec(ctx, q, userID, symbol, reason)
	if err != nil {
		return fmt.Errorf("mark latest agent run stopped: %w", err)
	}
	return nil
}
