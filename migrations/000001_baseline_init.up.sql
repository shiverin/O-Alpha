CREATE OR REPLACE FUNCTION set_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT DEFAULT 'user' NOT NULL,
    display_name TEXT NOT NULL,
    subscription_tier TEXT DEFAULT 'free' NOT NULL,
    base_currency TEXT DEFAULT 'USD' NOT NULL,
    timezone TEXT DEFAULT 'UTC' NOT NULL,
    is_onboarded BOOLEAN DEFAULT false NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    CONSTRAINT chk_users_role CHECK (role IN ('user', 'admin')),
    CONSTRAINT chk_users_subscription_tier CHECK (subscription_tier IN ('free', 'pro', 'enterprise')),
    CONSTRAINT chk_users_currency CHECK (char_length(base_currency) = 3)
);

CREATE UNIQUE INDEX users_username_lower_idx ON users (lower(username));

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_column();

CREATE TABLE sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    token_hash TEXT UNIQUE NOT NULL,
    user_agent TEXT,
    ip_address INET,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE INDEX sessions_user_expires_idx ON sessions (user_id, expires_at DESC);
CREATE INDEX sessions_active_idx ON sessions (user_id, expires_at DESC) WHERE revoked_at IS NULL;

CREATE TRIGGER trg_sessions_updated_at
    BEFORE UPDATE ON sessions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_column();

CREATE TABLE assets (
    symbol TEXT PRIMARY KEY,
    name TEXT DEFAULT '' NOT NULL,
    asset_class TEXT DEFAULT 'equity' NOT NULL,
    exchange TEXT DEFAULT '' NOT NULL,
    currency TEXT DEFAULT 'USD' NOT NULL,
    active BOOLEAN DEFAULT true NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    CONSTRAINT chk_assets_symbol_upper CHECK (symbol = upper(symbol)),
    CONSTRAINT chk_assets_asset_class CHECK (asset_class IN ('equity', 'crypto', 'forex', 'fixed_income', 'cash', 'index', 'fund', 'option', 'future')),
    CONSTRAINT chk_assets_currency CHECK (char_length(currency) = 3)
);

CREATE INDEX assets_class_active_idx ON assets (asset_class, active);

CREATE TRIGGER trg_assets_updated_at
    BEFORE UPDATE ON assets
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_column();

CREATE TABLE accounts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    account_type TEXT DEFAULT 'paper' NOT NULL,
    provider TEXT DEFAULT 'internal' NOT NULL,
    provider_account_id TEXT DEFAULT '' NOT NULL,
    currency TEXT DEFAULT 'USD' NOT NULL,
    initial_cash NUMERIC(18, 4) DEFAULT 100000.0000 NOT NULL,
    cash_balance NUMERIC(18, 4) DEFAULT 100000.0000 NOT NULL,
    realized_pnl NUMERIC(18, 6) DEFAULT 0 NOT NULL,
    status TEXT DEFAULT 'active' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    CONSTRAINT chk_accounts_type CHECK (account_type IN ('paper', 'live')),
    CONSTRAINT chk_accounts_status CHECK (status IN ('active', 'paused', 'closed')),
    CONSTRAINT chk_accounts_currency CHECK (char_length(currency) = 3),
    CONSTRAINT chk_accounts_initial_cash CHECK (initial_cash >= 0),
    CONSTRAINT chk_accounts_cash_balance CHECK (cash_balance >= 0),
    UNIQUE (id, user_id),
    UNIQUE (user_id, account_type, provider, provider_account_id)
);

CREATE INDEX accounts_user_status_idx ON accounts (user_id, status);

CREATE TRIGGER trg_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_column();

CREATE TABLE agent_settings (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE PRIMARY KEY,
    risk_profile TEXT DEFAULT 'moderate' NOT NULL,
    leverage INT DEFAULT 1 NOT NULL,
    max_positions INT DEFAULT 5 NOT NULL,
    stop_loss_pct NUMERIC(6, 2) DEFAULT 2.00 NOT NULL,
    take_profit_pct NUMERIC(6, 2) DEFAULT 4.00 NOT NULL,
    rebalance_freq TEXT DEFAULT 'daily' NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    CONSTRAINT chk_agent_settings_risk_profile CHECK (risk_profile IN ('conservative', 'moderate', 'aggressive')),
    CONSTRAINT chk_agent_settings_rebalance_freq CHECK (rebalance_freq IN ('hourly', 'daily', 'weekly', 'monthly')),
    CONSTRAINT chk_agent_settings_leverage CHECK (leverage BETWEEN 1 AND 10),
    CONSTRAINT chk_agent_settings_max_positions CHECK (max_positions BETWEEN 1 AND 100),
    CONSTRAINT chk_agent_settings_stop_loss CHECK (stop_loss_pct > 0 AND stop_loss_pct <= 100),
    CONSTRAINT chk_agent_settings_take_profit CHECK (take_profit_pct > 0 AND take_profit_pct <= 100)
);

CREATE TRIGGER trg_agent_settings_updated_at
    BEFORE UPDATE ON agent_settings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_column();

CREATE TABLE strategy_configs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    strategy_type TEXT NOT NULL,
    version INT DEFAULT 1 NOT NULL,
    visibility TEXT DEFAULT 'private' NOT NULL,
    parameters JSONB DEFAULT '{}'::jsonb NOT NULL,
    is_active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    CONSTRAINT chk_strategy_configs_type CHECK (strategy_type IN ('MA_CROSSOVER', 'KALMAN', 'REGIME_DETECTOR', 'PAIRS')),
    CONSTRAINT chk_strategy_configs_visibility CHECK (visibility IN ('private', 'shared', 'system')),
    CONSTRAINT chk_strategy_configs_version CHECK (version > 0),
    UNIQUE (user_id, name, version)
);

CREATE INDEX strategy_configs_user_active_idx ON strategy_configs (user_id, is_active, updated_at DESC);
CREATE INDEX strategy_configs_parameters_gin_idx ON strategy_configs USING GIN (parameters);

CREATE TRIGGER trg_strategy_configs_updated_at
    BEFORE UPDATE ON strategy_configs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_column();

CREATE TABLE backtest_runs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    strategy_config_id BIGINT REFERENCES strategy_configs(id) ON DELETE SET NULL,
    strategy_type TEXT NOT NULL,
    symbol TEXT REFERENCES assets(symbol) ON DELETE RESTRICT NOT NULL,
    timeframe TEXT NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    initial_cash NUMERIC(18, 4) NOT NULL,
    final_equity NUMERIC(18, 4) NOT NULL,
    total_return_pct NUMERIC(12, 6) NOT NULL,
    annual_return_pct NUMERIC(12, 6) DEFAULT 0 NOT NULL,
    sharpe_ratio NUMERIC(12, 6) NOT NULL,
    sortino_ratio NUMERIC(12, 6) NOT NULL,
    max_drawdown_pct NUMERIC(12, 6) NOT NULL,
    num_trades INT NOT NULL,
    parameters JSONB DEFAULT '{}'::jsonb NOT NULL,
    equity_curve JSONB NOT NULL,
    diagnostics JSONB DEFAULT '{}'::jsonb NOT NULL,
    status TEXT DEFAULT 'completed' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    CONSTRAINT chk_backtest_runs_time CHECK (end_time > start_time),
    CONSTRAINT chk_backtest_runs_cash CHECK (initial_cash > 0),
    CONSTRAINT chk_backtest_runs_strategy_type CHECK (strategy_type IN ('MA_CROSSOVER', 'KALMAN', 'REGIME_DETECTOR', 'PAIRS')),
    CONSTRAINT chk_backtest_runs_status CHECK (status IN ('queued', 'running', 'completed', 'failed'))
);

CREATE INDEX backtest_runs_user_created_idx ON backtest_runs (user_id, created_at DESC);
CREATE INDEX backtest_runs_symbol_timeframe_idx ON backtest_runs (symbol, timeframe, created_at DESC);

CREATE TABLE backtest_trades (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    backtest_run_id BIGINT REFERENCES backtest_runs(id) ON DELETE CASCADE NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    action TEXT NOT NULL,
    symbol TEXT REFERENCES assets(symbol) ON DELETE RESTRICT NOT NULL,
    price NUMERIC(18, 6) NOT NULL,
    qty NUMERIC(20, 8) NOT NULL,
    realized_pnl NUMERIC(18, 6) DEFAULT 0 NOT NULL,

    CONSTRAINT chk_backtest_trades_action CHECK (action IN ('BUY_LONG', 'SELL_LONG', 'SELL_SHORT', 'COVER_SHORT')),
    CONSTRAINT chk_backtest_trades_price CHECK (price >= 0),
    CONSTRAINT chk_backtest_trades_qty CHECK (qty > 0)
);

CREATE INDEX backtest_trades_run_time_idx ON backtest_trades (backtest_run_id, timestamp);

CREATE TABLE agent_runs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    strategy_config_id BIGINT REFERENCES strategy_configs(id) ON DELETE SET NULL,
    symbol TEXT REFERENCES assets(symbol) ON DELETE RESTRICT NOT NULL,
    strategy_type TEXT NOT NULL,
    timeframe TEXT NOT NULL,
    mode TEXT DEFAULT 'paper' NOT NULL,
    status TEXT DEFAULT 'starting' NOT NULL,
    initial_cash NUMERIC(18, 4) NOT NULL,
    use_websocket BOOLEAN DEFAULT false NOT NULL,
    parameters JSONB DEFAULT '{}'::jsonb NOT NULL,
    started_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    stopped_at TIMESTAMPTZ,
    last_heartbeat_at TIMESTAMPTZ,
    stop_reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    FOREIGN KEY (account_id, user_id) REFERENCES accounts(id, user_id) ON DELETE CASCADE,
    CONSTRAINT chk_agent_runs_strategy_type CHECK (strategy_type IN ('MA_CROSSOVER', 'KALMAN', 'REGIME_DETECTOR', 'PAIRS')),
    CONSTRAINT chk_agent_runs_mode CHECK (mode IN ('paper', 'live')),
    CONSTRAINT chk_agent_runs_status CHECK (status IN ('starting', 'running', 'stopping', 'stopped', 'failed')),
    CONSTRAINT chk_agent_runs_initial_cash CHECK (initial_cash > 0)
);

CREATE INDEX agent_runs_user_status_idx ON agent_runs (user_id, status, started_at DESC);
CREATE INDEX agent_runs_account_symbol_idx ON agent_runs (account_id, symbol, started_at DESC);

CREATE TRIGGER trg_agent_runs_updated_at
    BEFORE UPDATE ON agent_runs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_column();

CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    agent_run_id BIGINT REFERENCES agent_runs(id) ON DELETE SET NULL,
    strategy_config_id BIGINT REFERENCES strategy_configs(id) ON DELETE SET NULL,
    symbol TEXT REFERENCES assets(symbol) ON DELETE RESTRICT NOT NULL,
    side TEXT NOT NULL,
    position_side TEXT DEFAULT 'long' NOT NULL,
    order_type TEXT DEFAULT 'market' NOT NULL,
    time_in_force TEXT DEFAULT 'day' NOT NULL,
    qty NUMERIC(20, 8) NOT NULL,
    limit_price NUMERIC(18, 6),
    stop_price NUMERIC(18, 6),
    status TEXT DEFAULT 'new' NOT NULL,
    client_order_id TEXT,
    provider_order_id TEXT,
    submitted_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    filled_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    rejected_reason TEXT,
    metadata JSONB DEFAULT '{}'::jsonb NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    FOREIGN KEY (account_id, user_id) REFERENCES accounts(id, user_id) ON DELETE CASCADE,
    CONSTRAINT chk_orders_side CHECK (side IN ('buy', 'sell')),
    CONSTRAINT chk_orders_position_side CHECK (position_side IN ('long', 'short')),
    CONSTRAINT chk_orders_type CHECK (order_type IN ('market', 'limit', 'stop', 'stop_limit')),
    CONSTRAINT chk_orders_tif CHECK (time_in_force IN ('day', 'gtc', 'ioc', 'fok')),
    CONSTRAINT chk_orders_status CHECK (status IN ('new', 'accepted', 'partially_filled', 'filled', 'canceled', 'rejected', 'error')),
    CONSTRAINT chk_orders_qty CHECK (qty > 0),
    CONSTRAINT chk_orders_limit_price CHECK (limit_price IS NULL OR limit_price >= 0),
    CONSTRAINT chk_orders_stop_price CHECK (stop_price IS NULL OR stop_price >= 0)
);

CREATE UNIQUE INDEX orders_client_order_id_idx ON orders (client_order_id) WHERE client_order_id IS NOT NULL;
CREATE INDEX orders_user_status_idx ON orders (user_id, status, submitted_at DESC);
CREATE INDEX orders_account_symbol_idx ON orders (account_id, symbol, submitted_at DESC);

CREATE TRIGGER trg_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_column();

CREATE TABLE fills (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id BIGINT REFERENCES orders(id) ON DELETE CASCADE NOT NULL,
    user_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    symbol TEXT REFERENCES assets(symbol) ON DELETE RESTRICT NOT NULL,
    side TEXT NOT NULL,
    position_side TEXT DEFAULT 'long' NOT NULL,
    price NUMERIC(18, 6) NOT NULL,
    qty NUMERIC(20, 8) NOT NULL,
    slippage NUMERIC(10, 6) DEFAULT 0 NOT NULL,
    commission NUMERIC(18, 6) DEFAULT 0 NOT NULL,
    liquidity TEXT DEFAULT 'unknown' NOT NULL,
    filled_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    FOREIGN KEY (account_id, user_id) REFERENCES accounts(id, user_id) ON DELETE CASCADE,
    CONSTRAINT chk_fills_side CHECK (side IN ('buy', 'sell')),
    CONSTRAINT chk_fills_position_side CHECK (position_side IN ('long', 'short')),
    CONSTRAINT chk_fills_price CHECK (price >= 0),
    CONSTRAINT chk_fills_qty CHECK (qty > 0),
    CONSTRAINT chk_fills_commission CHECK (commission >= 0),
    CONSTRAINT chk_fills_liquidity CHECK (liquidity IN ('maker', 'taker', 'unknown'))
);

CREATE INDEX fills_user_time_idx ON fills (user_id, filled_at DESC);
CREATE INDEX fills_account_symbol_idx ON fills (account_id, symbol, filled_at DESC);

CREATE TABLE positions (
    user_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    symbol TEXT REFERENCES assets(symbol) ON DELETE RESTRICT NOT NULL,
    position_side TEXT DEFAULT 'long' NOT NULL,
    qty NUMERIC(20, 8) NOT NULL,
    avg_entry_price NUMERIC(18, 6) NOT NULL,
    current_price NUMERIC(18, 6) NOT NULL,
    realized_pnl NUMERIC(18, 6) DEFAULT 0 NOT NULL,
    opened_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    PRIMARY KEY (account_id, symbol, position_side),
    FOREIGN KEY (account_id, user_id) REFERENCES accounts(id, user_id) ON DELETE CASCADE,
    CONSTRAINT chk_positions_side CHECK (position_side IN ('long', 'short')),
    CONSTRAINT chk_positions_qty CHECK (qty >= 0),
    CONSTRAINT chk_positions_avg_entry CHECK (avg_entry_price >= 0),
    CONSTRAINT chk_positions_current_price CHECK (current_price >= 0)
);

CREATE INDEX positions_user_symbol_idx ON positions (user_id, symbol);

CREATE TRIGGER trg_positions_updated_at
    BEFORE UPDATE ON positions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_column();

CREATE TABLE cash_ledger (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    event_type TEXT NOT NULL,
    amount NUMERIC(18, 6) NOT NULL,
    currency TEXT DEFAULT 'USD' NOT NULL,
    balance_after NUMERIC(18, 6) NOT NULL,
    related_order_id BIGINT REFERENCES orders(id) ON DELETE SET NULL,
    related_fill_id BIGINT REFERENCES fills(id) ON DELETE SET NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    FOREIGN KEY (account_id, user_id) REFERENCES accounts(id, user_id) ON DELETE CASCADE,
    CONSTRAINT chk_cash_ledger_event CHECK (event_type IN ('initial_deposit', 'deposit', 'withdrawal', 'trade_buy', 'trade_sell', 'commission', 'adjustment')),
    CONSTRAINT chk_cash_ledger_currency CHECK (char_length(currency) = 3),
    CONSTRAINT chk_cash_ledger_balance CHECK (balance_after >= 0)
);

CREATE INDEX cash_ledger_account_time_idx ON cash_ledger (account_id, created_at DESC);

CREATE TABLE system_alerts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    account_id BIGINT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    alert_type TEXT NOT NULL,
    source TEXT DEFAULT 'system' NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    resolved_at TIMESTAMPTZ,

    FOREIGN KEY (account_id, user_id) REFERENCES accounts(id, user_id) ON DELETE CASCADE,
    CONSTRAINT chk_alert_type CHECK (alert_type IN ('INFO', 'WARNING', 'CRITICAL'))
);

CREATE INDEX alerts_user_time_idx ON system_alerts (user_id, created_at DESC);
CREATE INDEX alerts_unresolved_idx ON system_alerts (user_id, alert_type, created_at DESC) WHERE resolved_at IS NULL;

CREATE TABLE portfolio_snapshots (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    cash_value NUMERIC(18, 6) DEFAULT 0 NOT NULL,
    positions_value NUMERIC(18, 6) DEFAULT 0 NOT NULL,
    total_asset_value NUMERIC(18, 6) DEFAULT 0 NOT NULL,
    realized_pnl NUMERIC(18, 6) DEFAULT 0 NOT NULL,
    unrealized_pnl NUMERIC(18, 6) DEFAULT 0 NOT NULL,
    change_percent_24h NUMERIC(8, 4) DEFAULT 0 NOT NULL,
    change_dollar_24h NUMERIC(18, 6) DEFAULT 0 NOT NULL,
    estimated_annual_yield NUMERIC(18, 6) DEFAULT 0 NOT NULL,
    target_progress_percent NUMERIC(8, 4) DEFAULT 0 NOT NULL,

    FOREIGN KEY (account_id, user_id) REFERENCES accounts(id, user_id) ON DELETE CASCADE,
    CONSTRAINT chk_portfolio_snapshot_cash CHECK (cash_value >= 0),
    CONSTRAINT chk_portfolio_snapshot_positions CHECK (positions_value >= 0),
    CONSTRAINT chk_portfolio_snapshot_total CHECK (total_asset_value >= 0)
);

CREATE INDEX portfolio_snapshots_user_time_idx ON portfolio_snapshots (user_id, timestamp DESC);
CREATE INDEX portfolio_snapshots_account_time_idx ON portfolio_snapshots (account_id, timestamp DESC);

CREATE TABLE bars (
    time TIMESTAMPTZ NOT NULL,
    symbol TEXT REFERENCES assets(symbol) ON DELETE RESTRICT NOT NULL,
    timeframe TEXT NOT NULL,
    open DOUBLE PRECISION NOT NULL,
    high DOUBLE PRECISION NOT NULL,
    low DOUBLE PRECISION NOT NULL,
    close DOUBLE PRECISION NOT NULL,
    volume BIGINT NOT NULL,
    trade_count BIGINT,
    vwap DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    PRIMARY KEY (time, symbol, timeframe),

    CONSTRAINT chk_bars_symbol_upper CHECK (symbol = upper(symbol)),
    CONSTRAINT chk_bars_timeframe CHECK (timeframe IN ('1Min', '5Min', '15Min', '1Hour', '1Day')),
    CONSTRAINT chk_bars_prices_nonnegative CHECK (open >= 0 AND high >= 0 AND low >= 0 AND close >= 0),
    CONSTRAINT chk_bars_ohlc_consistent CHECK (high >= low AND high >= open AND high >= close AND low <= open AND low <= close),
    CONSTRAINT chk_bars_volume_nonnegative CHECK (volume >= 0),
    CONSTRAINT chk_bars_trade_count_nonnegative CHECK (trade_count IS NULL OR trade_count >= 0),
    CONSTRAINT chk_bars_vwap_nonnegative CHECK (vwap IS NULL OR vwap >= 0)
) PARTITION BY RANGE (time);

CREATE INDEX bars_symbol_timeframe_time_idx ON bars (symbol, timeframe, time DESC);

CREATE TABLE bars_y2024 PARTITION OF bars FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');
CREATE TABLE bars_y2025 PARTITION OF bars FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');
CREATE TABLE bars_y2026 PARTITION OF bars FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');
CREATE TABLE bars_default PARTITION OF bars DEFAULT;
