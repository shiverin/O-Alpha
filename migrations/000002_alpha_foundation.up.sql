ALTER TABLE assets
    ADD COLUMN IF NOT EXISTS sector TEXT,
    ADD COLUMN IF NOT EXISTS industry TEXT,
    ADD COLUMN IF NOT EXISTS is_etf BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS is_leveraged_etf BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS is_inverse_etf BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE bars
    ADD COLUMN IF NOT EXISTS feed TEXT NOT NULL DEFAULT 'iex',
    ADD COLUMN IF NOT EXISTS adjustment TEXT NOT NULL DEFAULT 'raw',
    ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'alpaca';

ALTER TABLE bars DROP CONSTRAINT IF EXISTS bars_pkey;
DROP INDEX IF EXISTS bars_time_symbol_timeframe_key;
DROP INDEX IF EXISTS bars_symbol_timeframe_time_idx;

CREATE UNIQUE INDEX IF NOT EXISTS bars_unique_dataset_idx
ON bars (time, symbol, timeframe, feed, adjustment, source);

CREATE INDEX IF NOT EXISTS bars_symbol_timeframe_dataset_time_idx
ON bars (symbol, timeframe, feed, adjustment, source, time DESC);

CREATE TABLE IF NOT EXISTS universes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE TABLE IF NOT EXISTS universe_members (
    universe_id BIGINT NOT NULL REFERENCES universes(id) ON DELETE CASCADE,
    symbol TEXT NOT NULL REFERENCES assets(symbol) ON DELETE RESTRICT,
    weight_cap DOUBLE PRECISION,
    sector TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (universe_id, symbol)
);

CREATE TABLE IF NOT EXISTS portfolio_backtest_runs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    strategy_type TEXT NOT NULL,
    symbols TEXT[] NOT NULL,
    timeframe TEXT NOT NULL,
    feed TEXT NOT NULL,
    adjustment TEXT NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    initial_cash DOUBLE PRECISION NOT NULL,
    final_equity DOUBLE PRECISION NOT NULL,
    metrics JSONB NOT NULL,
    parameters JSONB NOT NULL,
    equity_curve JSONB NOT NULL,
    trades JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sleeve_returns (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    run_id BIGINT NOT NULL,
    sleeve TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    return_value DOUBLE PRECISION NOT NULL,
    gross_exposure DOUBLE PRECISION,
    net_exposure DOUBLE PRECISION,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS sleeve_returns_run_sleeve_time_idx
ON sleeve_returns (run_id, sleeve, timestamp);

CREATE TABLE IF NOT EXISTS ml_model_artifacts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    model_name TEXT NOT NULL,
    model_type TEXT NOT NULL,
    strategy_scope TEXT NOT NULL,
    artifact_uri TEXT NOT NULL,
    feature_spec JSONB NOT NULL,
    label_config JSONB NOT NULL,
    training_config JSONB NOT NULL,
    train_start TIMESTAMPTZ NOT NULL,
    train_end TIMESTAMPTZ NOT NULL,
    validation_start TIMESTAMPTZ,
    validation_end TIMESTAMPTZ,
    auc DOUBLE PRECISION,
    logloss DOUBLE PRECISION,
    sharpe_net DOUBLE PRECISION,
    sortino_net DOUBLE PRECISION,
    max_drawdown_pct DOUBLE PRECISION,
    dsr DOUBLE PRECISION,
    pbo DOUBLE PRECISION,
    leaves_parity_max_abs_error DOUBLE PRECISION,
    leaves_parity_passed BOOLEAN NOT NULL DEFAULT false,
    status TEXT NOT NULL DEFAULT 'candidate',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ml_model_artifacts_promoted_idx
ON ml_model_artifacts (model_name, status, created_at DESC);

CREATE TABLE IF NOT EXISTS pair_candidates (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    symbol_y TEXT NOT NULL REFERENCES assets(symbol) ON DELETE RESTRICT,
    symbol_x TEXT NOT NULL REFERENCES assets(symbol) ON DELETE RESTRICT,
    timeframe TEXT NOT NULL,
    formation_start TIMESTAMPTZ NOT NULL,
    formation_end TIMESTAMPTZ NOT NULL,
    correlation DOUBLE PRECISION,
    engle_granger_pvalue DOUBLE PRECISION,
    johansen_trace_stat DOUBLE PRECISION,
    half_life_bars DOUBLE PRECISION,
    hurst DOUBLE PRECISION,
    avg_spread_bps DOUBLE PRECISION,
    estimated_round_trip_cost_bps DOUBLE PRECISION,
    approved BOOLEAN NOT NULL DEFAULT false,
    status TEXT NOT NULL DEFAULT 'candidate',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS pair_candidates_approved_idx
ON pair_candidates (approved, status, formation_end DESC);

CREATE TABLE IF NOT EXISTS strategy_trials (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    family TEXT NOT NULL,
    config_hash TEXT NOT NULL,
    parameters JSONB NOT NULL,
    train_period TSTZRANGE,
    test_period TSTZRANGE,
    sharpe DOUBLE PRECISION,
    sortino DOUBLE PRECISION,
    calmar DOUBLE PRECISION,
    max_drawdown DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE strategy_configs DROP CONSTRAINT IF EXISTS chk_strategy_configs_type;
ALTER TABLE strategy_configs ADD CONSTRAINT chk_strategy_configs_type
CHECK (strategy_type IN ('MA_CROSSOVER', 'KALMAN', 'REGIME_DETECTOR', 'PAIRS', 'HMM_ENSEMBLE', 'ML_META_LABEL', 'XSEC_MOMENTUM', 'KALMAN_COINTEGRATION', 'MULTI_ENGINE_ENSEMBLE'));

ALTER TABLE backtest_runs DROP CONSTRAINT IF EXISTS chk_backtest_runs_strategy_type;
ALTER TABLE backtest_runs ADD CONSTRAINT chk_backtest_runs_strategy_type
CHECK (strategy_type IN ('MA_CROSSOVER', 'KALMAN', 'REGIME_DETECTOR', 'PAIRS', 'HMM_ENSEMBLE', 'ML_META_LABEL', 'XSEC_MOMENTUM', 'KALMAN_COINTEGRATION', 'MULTI_ENGINE_ENSEMBLE'));

ALTER TABLE agent_runs DROP CONSTRAINT IF EXISTS chk_agent_runs_strategy_type;
ALTER TABLE agent_runs ADD CONSTRAINT chk_agent_runs_strategy_type
CHECK (strategy_type IN ('MA_CROSSOVER', 'KALMAN', 'REGIME_DETECTOR', 'PAIRS', 'HMM_ENSEMBLE', 'ML_META_LABEL', 'XSEC_MOMENTUM', 'KALMAN_COINTEGRATION', 'MULTI_ENGINE_ENSEMBLE'));
