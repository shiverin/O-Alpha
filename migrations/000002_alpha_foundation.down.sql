ALTER TABLE agent_runs DROP CONSTRAINT IF EXISTS chk_agent_runs_strategy_type;
ALTER TABLE agent_runs ADD CONSTRAINT chk_agent_runs_strategy_type
CHECK (strategy_type IN ('MA_CROSSOVER', 'KALMAN', 'REGIME_DETECTOR', 'PAIRS'));

ALTER TABLE backtest_runs DROP CONSTRAINT IF EXISTS chk_backtest_runs_strategy_type;
ALTER TABLE backtest_runs ADD CONSTRAINT chk_backtest_runs_strategy_type
CHECK (strategy_type IN ('MA_CROSSOVER', 'KALMAN', 'REGIME_DETECTOR', 'PAIRS'));

ALTER TABLE strategy_configs DROP CONSTRAINT IF EXISTS chk_strategy_configs_type;
ALTER TABLE strategy_configs ADD CONSTRAINT chk_strategy_configs_type
CHECK (strategy_type IN ('MA_CROSSOVER', 'KALMAN', 'REGIME_DETECTOR', 'PAIRS'));

DROP TABLE IF EXISTS strategy_trials;
DROP INDEX IF EXISTS pair_candidates_approved_idx;
DROP TABLE IF EXISTS pair_candidates;
DROP INDEX IF EXISTS ml_model_artifacts_promoted_idx;
DROP TABLE IF EXISTS ml_model_artifacts;
DROP INDEX IF EXISTS sleeve_returns_run_sleeve_time_idx;
DROP TABLE IF EXISTS sleeve_returns;
DROP TABLE IF EXISTS portfolio_backtest_runs;
DROP TABLE IF EXISTS universe_members;
DROP TABLE IF EXISTS universes;

DROP INDEX IF EXISTS bars_symbol_timeframe_dataset_time_idx;
DROP INDEX IF EXISTS bars_unique_dataset_idx;

ALTER TABLE bars
    DROP COLUMN IF EXISTS source,
    DROP COLUMN IF EXISTS adjustment,
    DROP COLUMN IF EXISTS feed;

ALTER TABLE bars ADD PRIMARY KEY (time, symbol, timeframe);
CREATE INDEX IF NOT EXISTS bars_symbol_timeframe_time_idx ON bars (symbol, timeframe, time DESC);

ALTER TABLE assets
    DROP COLUMN IF EXISTS is_inverse_etf,
    DROP COLUMN IF EXISTS is_leveraged_etf,
    DROP COLUMN IF EXISTS is_etf,
    DROP COLUMN IF EXISTS industry,
    DROP COLUMN IF EXISTS sector;
