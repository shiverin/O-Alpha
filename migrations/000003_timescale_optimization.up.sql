-- 1. Add timeframe column
ALTER TABLE bars ADD COLUMN timeframe TEXT DEFAULT '1Day' NOT NULL;

-- 2. Rename the old table so we can create a partitioned replacement
ALTER TABLE bars RENAME TO bars_old;

-- 3. Create the new partitioned table (partitioned by the 'time' column)
CREATE TABLE bars (
    time       TIMESTAMPTZ NOT NULL,
    symbol     TEXT NOT NULL,
    timeframe  TEXT NOT NULL,
    open       DOUBLE PRECISION NOT NULL,
    high       DOUBLE PRECISION NOT NULL,
    low        DOUBLE PRECISION NOT NULL,
    close      DOUBLE PRECISION NOT NULL,
    volume     BIGINT NOT NULL,
    PRIMARY KEY (time, symbol, timeframe)
) PARTITION BY RANGE (time);

-- 4. Create initial partitions for the current year (you can automate future ones with pg_partman later)
CREATE TABLE bars_y2025 PARTITION OF bars FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');
CREATE TABLE bars_y2026 PARTITION OF bars FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');

-- 5. Move data from the old table to the partitioned table
INSERT INTO bars (time, symbol, timeframe, open, high, low, close, volume)
SELECT time, symbol, '1Day', open, high, low, close, volume FROM bars_old;

-- 6. Drop the old table
DROP TABLE bars_old;