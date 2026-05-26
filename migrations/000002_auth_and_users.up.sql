-- 1. Users Table (Add IF NOT EXISTS)
CREATE TABLE IF NOT EXISTS users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT DEFAULT 'user' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- 2. User Sessions (Add IF NOT EXISTS)
CREATE TABLE IF NOT EXISTS user_sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    refresh_token TEXT UNIQUE NOT NULL,
    user_agent TEXT,
    ip_address TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Add IF NOT EXISTS for the index
CREATE INDEX IF NOT EXISTS user_sessions_user_id_idx ON user_sessions(user_id);

-- 3. Agent Settings / Preferences (Add IF NOT EXISTS)
CREATE TABLE IF NOT EXISTS agent_settings (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE PRIMARY KEY,
    risk_profile TEXT DEFAULT 'moderate' NOT NULL,
    leverage INT DEFAULT 1 NOT NULL,
    max_positions INT DEFAULT 5 NOT NULL,
    stop_loss_pct NUMERIC(5,2) DEFAULT 2.0 NOT NULL,
    take_profit_pct NUMERIC(5,2) DEFAULT 4.0 NOT NULL,
    rebalance_freq TEXT DEFAULT 'daily' NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);