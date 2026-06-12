package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type AgentRunSummary struct {
	ID              int64                  `json:"id"`
	Symbol          string                 `json:"symbol"`
	StrategyType    string                 `json:"strategy_type"`
	StrategyKey     string                 `json:"strategy_key,omitempty"`
	Timeframe       string                 `json:"timeframe"`
	Mode            string                 `json:"mode"`
	Status          string                 `json:"status"`
	InitialCash     float64                `json:"initial_cash"`
	Parameters      map[string]interface{} `json:"parameters,omitempty"`
	RuntimeState    map[string]interface{} `json:"runtime_state,omitempty"`
	StartedAt       time.Time              `json:"started_at"`
	LastHeartbeatAt *time.Time             `json:"last_heartbeat_at,omitempty"`
}

func (r *AgentRepository) ListActiveAgentRuns(ctx context.Context, userID int64) ([]AgentRunSummary, error) {
	const q = `
		SELECT id, symbol, strategy_type, timeframe, mode, status, initial_cash, parameters, started_at, last_heartbeat_at
		FROM agent_runs
		WHERE user_id = $1 AND status IN ('starting', 'running')
		ORDER BY started_at DESC`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list active agent runs: %w", err)
	}
	defer rows.Close()

	summaries := make([]AgentRunSummary, 0)
	for rows.Next() {
		var s AgentRunSummary
		var paramsBytes []byte
		if err := rows.Scan(
			&s.ID,
			&s.Symbol,
			&s.StrategyType,
			&s.Timeframe,
			&s.Mode,
			&s.Status,
			&s.InitialCash,
			&paramsBytes,
			&s.StartedAt,
			&s.LastHeartbeatAt,
		); err != nil {
			return nil, err
		}
		if len(paramsBytes) > 0 {
			if err := json.Unmarshal(paramsBytes, &s.Parameters); err != nil {
				return nil, fmt.Errorf("unmarshal agent run parameters: %w", err)
			}
			if key, ok := s.Parameters["strategy_key"].(string); ok {
				s.StrategyKey = key
			}
			if state, ok := s.Parameters["runtime_state"].(map[string]interface{}); ok {
				s.RuntimeState = state
			}
		}
		summaries = append(summaries, s)
	}
	return summaries, rows.Err()
}

func (r *AgentRepository) UpdateAgentRunHeartbeat(ctx context.Context, runID int64) error {
	const q = `UPDATE agent_runs SET last_heartbeat_at = NOW() WHERE id = $1 AND status = 'running'`
	if _, err := r.db.Exec(ctx, q, runID); err != nil {
		return fmt.Errorf("update agent run heartbeat: %w", err)
	}
	return nil
}

func (r *AgentRepository) UpdateAgentRunRuntimeState(ctx context.Context, runID int64, runtimeState map[string]interface{}) error {
	if runtimeState == nil {
		runtimeState = map[string]interface{}{}
	}
	stateBytes, err := json.Marshal(map[string]interface{}{
		"runtime_state": runtimeState,
	})
	if err != nil {
		return fmt.Errorf("marshal agent runtime state: %w", err)
	}

	const q = `
		UPDATE agent_runs
		SET parameters = parameters || $2::jsonb
		WHERE id = $1 AND status IN ('starting', 'running')`
	if _, err := r.db.Exec(ctx, q, runID, stateBytes); err != nil {
		return fmt.Errorf("update agent runtime state: %w", err)
	}
	return nil
}

func (r *AgentRepository) MarkActivePortfolioRunStopped(ctx context.Context, userID int64, reason string) error {
	const q = `
		WITH latest AS (
			SELECT id
			FROM agent_runs
			WHERE user_id = $1
				AND strategy_type = 'PORTFOLIO_CATALOG'
				AND status IN ('starting', 'running', 'stopping')
			ORDER BY started_at DESC
			LIMIT 1
		)
		UPDATE agent_runs
		SET status = 'stopped',
			stopped_at = NOW(),
			stop_reason = $2
		WHERE id = (SELECT id FROM latest)`
	if _, err := r.db.Exec(ctx, q, userID, reason); err != nil {
		return fmt.Errorf("mark active portfolio run stopped: %w", err)
	}
	return nil
}

func (r *AgentRepository) MarkOrphanedAgentRunsFailed(ctx context.Context, staleAfter time.Duration) (int64, error) {
	const q = `
		UPDATE agent_runs
		SET status = 'failed',
			stopped_at = NOW(),
			stop_reason = 'orphaned_on_restart'
		WHERE status IN ('starting', 'running', 'stopping')
			AND (
				$1::bigint = 0
				OR last_heartbeat_at IS NULL
				OR last_heartbeat_at < NOW() - ($1::text || ' seconds')::interval
			)`
	seconds := int64(staleAfter.Seconds())
	tag, err := r.db.Exec(ctx, q, seconds)
	if err != nil {
		return 0, fmt.Errorf("mark orphaned agent runs failed: %w", err)
	}
	return tag.RowsAffected(), nil
}
