-- 1. Create a temporary standard table matching the original 000001_init schema
CREATE TABLE bars_normal (
    time   TIMESTAMPTZ NOT NULL,
    symbol TEXT NOT NULL,
    open   DOUBLE PRECISION NOT NULL,
    high   DOUBLE PRECISION NOT NULL,
    low    DOUBLE PRECISION NOT NULL,
    close  DOUBLE PRECISION NOT NULL,
    volume BIGINT NOT NULL,
    PRIMARY KEY (time, symbol)
);

-- 2. Copy the data back from your partitioned layout 
-- (Omitting the timeframe column to revert to the exact original layout)
INSERT INTO bars_normal (time, symbol, open, high, low, close, volume)
SELECT time, symbol, open, high, low, close, volume FROM bars;

-- 3. Drop the partitioned parent table
-- This automatically safely cascades and drops the partition children (bars_y2025, bars_y2026, etc.)
DROP TABLE bars CASCADE;

-- 4. Rename the normal table back to 'bars'
ALTER TABLE bars_normal RENAME TO bars;