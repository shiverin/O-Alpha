package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AgentSettings maps directly to the schema definition inside your SQL migrations.
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

// AgentRepository isolates trading parameter persistence routines from user authentication layers.
type AgentRepository struct {
	db *pgxpool.Pool
}

// NewAgentRepository creates a new instance of the agent configuration data accessor.
func NewAgentRepository(db *pgxpool.Pool) *AgentRepository {
	return &AgentRepository{db: db}
}

// GetAgentSettings checks if a configuration blueprint exists for a target user ID.
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
		if errors.Is(err, pgx.ErrNoRows) { // High-performance sentinel checking
			return nil, nil
		}
		return nil, fmt.Errorf("select agent settings: %w", err)
	}
	return &s, nil
}

// SaveAgentSettings handles both creation and running parameter syncs using an atomic upsert statement.
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
