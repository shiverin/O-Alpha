-- =============================================================================
-- 01. GENERAL UTILITY CONFIGURATIONS & TRIGGER FUNCTIONS
-- =============================================================================
CREATE OR REPLACE FUNCTION set_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- 02. CORE USER IDENTITY & PROFILE ENTITY (MERGED)
-- =============================================================================
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username TEXT UNIQUE NOT NULL, 
    password_hash TEXT NOT NULL,
    role TEXT DEFAULT 'user' NOT NULL,
    
	-- Merged Profile Parameters
    display_name TEXT NOT NULL,
    subscription_tier INT DEFAULT 0 NOT NULL, 
    base_currency TEXT DEFAULT 'USD' NOT NULL,
    timezone TEXT DEFAULT 'UTC' NOT NULL,
    is_onboarded BOOLEAN DEFAULT false NOT NULL,
    
	created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    CONSTRAINT chk_subscription_tier CHECK (subscription_tier BETWEEN 0 AND 2)
);

CREATE TRIGGER trg_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_column();

-- =============================================================================
-- 03. USER PREFERENCE & AGENT OPERATIONS CONFIGURATIONS
-- =============================================================================
CREATE TABLE agent_settings (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE PRIMARY KEY,
    risk_profile TEXT DEFAULT 'moderate' NOT NULL,
    leverage INT DEFAULT 1 NOT NULL,
    max_positions INT DEFAULT 5 NOT NULL,
    stop_loss_pct NUMERIC(5,2) DEFAULT 2.0 NOT NULL,
    take_profit_pct NUMERIC(5,2) DEFAULT 4.0 NOT NULL,
    rebalance_freq TEXT DEFAULT 'daily' NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    
    CONSTRAINT chk_risk_profile CHECK (risk_profile IN ('conservative', 'moderate', 'aggressive')),
    CONSTRAINT chk_rebalance_freq CHECK (rebalance_freq IN ('hourly', 'daily', 'weekly', 'monthly'))
);

CREATE TRIGGER trg_agent_settings_updated_at 
    BEFORE UPDATE ON agent_settings 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_column();

-- =============================================================================
-- 04. STRATEGY RESEARCH & COMPUTE LOGGING INFRASTRUCTURE
-- =============================================================================
CREATE TABLE strategy_configs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    name TEXT NOT NULL,
    strategy_type TEXT NOT NULL,
    parameters JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE INDEX strategy_configs_user_id_idx ON strategy_configs(user_id);
CREATE INDEX strategy_configs_parameters_gin_idx ON strategy_configs USING GIN (parameters);

CREATE TRIGGER trg_strategy_configs_updated_at 
    BEFORE UPDATE ON strategy_configs 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_column();

CREATE TABLE backtest_runs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    strategy_config_id BIGINT REFERENCES strategy_configs(id) ON DELETE SET NULL,
    symbol TEXT NOT NULL,
    timeframe TEXT NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    initial_cash NUMERIC(15,2) NOT NULL,
    final_equity NUMERIC(15,2) NOT NULL,
    total_return_pct NUMERIC(10,4) NOT NULL,
    sharpe_ratio NUMERIC(8,4) NOT NULL,
    max_drawdown_pct NUMERIC(10,4) NOT NULL,
    num_trades INT NOT NULL,
    equity_curve JSONB NOT NULL, 
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE INDEX backtest_runs_user_id_idx ON backtest_runs(user_id);

-- =============================================================================
-- 05. REAL-TIME ACCOUNT TELEMETRY & AUDIT TRAILS
-- =============================================================================
CREATE TABLE trades (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    action TEXT NOT NULL,          
    symbol TEXT NOT NULL,          
    price NUMERIC(16, 4) NOT NULL, 
    qty NUMERIC(16, 8) NOT NULL,   
    slippage NUMERIC(6, 4) NOT NULL,
    status TEXT NOT NULL,
    
    CONSTRAINT chk_trade_action CHECK (action IN ('BUY_LONG', 'SELL_SHORT', 'SELL_LONG', 'COVER_SHORT')),
    CONSTRAINT chk_trade_status CHECK (status IN ('PENDING', 'FILLED', 'REJECTED', 'ERROR'))
);

CREATE INDEX trades_user_timestamp_idx ON trades (user_id, timestamp DESC);

CREATE TABLE positions (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    symbol TEXT NOT NULL,
    qty NUMERIC(16, 8) NOT NULL,
    avg_entry_price NUMERIC(16, 4) NOT NULL,
    current_price NUMERIC(16, 4) NOT NULL,
    PRIMARY KEY (user_id, symbol)
);

CREATE TABLE system_alerts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    alert_type TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    
    CONSTRAINT chk_alert_type CHECK (alert_type IN ('INFO', 'WARNING', 'CRITICAL'))
);

CREATE INDEX alerts_user_time_idx ON system_alerts (user_id, created_at DESC);

CREATE TABLE portfolio_snapshots (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    
    timestamp TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    total_asset_value NUMERIC(16, 4) DEFAULT 100000.0000 NOT NULL,
    change_percent_24h NUMERIC(5, 2) DEFAULT 0.00 NOT NULL,
    change_dollar_24h NUMERIC(16, 4) DEFAULT 0.0000 NOT NULL,
    estimated_annual_yield NUMERIC(16, 4) DEFAULT 0.0000 NOT NULL,
    target_progress_percent NUMERIC(5, 2) DEFAULT 0.00 NOT NULL,

    PRIMARY KEY (user_id, timestamp)
);

CREATE INDEX portfolio_snapshots_lookup_idx ON portfolio_snapshots (user_id, timestamp DESC);

-- =============================================================================
-- 06. NATIVE TIME-SERIES PARTITIONED MARKET DATA REGISTRIES
-- =============================================================================
CREATE TABLE bars (
    time       TIMESTAMPTZ NOT NULL,
    symbol     TEXT NOT NULL,
    timeframe  TEXT NOT NULL,
    open       DOUBLE PRECISION NOT NULL,
    high       DOUBLE PRECISION NOT NULL,
    low        DOUBLE PRECISION NOT NULL,
    close      DOUBLE PRECISION NOT NULL,
    volume     BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    PRIMARY KEY (time, symbol, timeframe)
) PARTITION BY RANGE (time);

CREATE INDEX bars_pruning_idx ON bars (symbol, timeframe, time DESC);

CREATE TABLE bars_y2024 PARTITION OF bars FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');
CREATE TABLE bars_y2025 PARTITION OF bars FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');
CREATE TABLE bars_y2026 PARTITION OF bars FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');
