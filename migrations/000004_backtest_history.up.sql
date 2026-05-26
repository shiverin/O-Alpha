-- 1. Saved Strategy Configurations
CREATE TABLE strategy_configs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    name TEXT NOT NULL,
    strategy_type TEXT NOT NULL, -- e.g., 'MA_CROSSOVER', 'REGIME_DETECTION'
    parameters JSONB NOT NULL,   -- e.g., {"fast_period": 10, "slow_period": 30}
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);
CREATE INDEX strategy_configs_user_id_idx ON strategy_configs(user_id);

-- 2. Backtest Runs (Cached Outputs)
CREATE TABLE backtest_runs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    strategy_config_id BIGINT REFERENCES strategy_configs(id) ON DELETE SET NULL,
    
    -- Inputs
    symbol TEXT NOT NULL,
    timeframe TEXT NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    initial_cash NUMERIC(15,2) NOT NULL,
    
    -- Cached Outputs (Metrics)
    final_equity NUMERIC(15,2) NOT NULL,
    total_return_pct NUMERIC(10,4) NOT NULL,
    sharpe_ratio NUMERIC(8,4) NOT NULL,
    max_drawdown_pct NUMERIC(10,4) NOT NULL,
    num_trades INT NOT NULL,
    
    -- Cached Outputs (Visual Data)
    -- Storing the equity curve array here as JSONB so the frontend can instantly render the Recharts component
    equity_curve JSONB NOT NULL, 
    
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);
CREATE INDEX backtest_runs_user_id_idx ON backtest_runs(user_id);
-- Index JSONB to easily filter runs by specific performance metrics if needed later
CREATE INDEX strategy_configs_parameters_idx ON strategy_configs USING GIN (parameters);