CREATE TABLE IF NOT EXISTS bars (
    time   TIMESTAMPTZ NOT NULL,
    symbol TEXT        NOT NULL,
    open   DOUBLE PRECISION NOT NULL,
    high   DOUBLE PRECISION NOT NULL,
    low    DOUBLE PRECISION NOT NULL,
    close  DOUBLE PRECISION NOT NULL,
    volume BIGINT      NOT NULL,
    PRIMARY KEY (time, symbol)
);

CREATE INDEX IF NOT EXISTS idx_bars_symbol_time ON bars (symbol, time DESC);
