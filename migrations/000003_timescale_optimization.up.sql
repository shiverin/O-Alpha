-- 1. Add timeframe column to the existing bars table safely (if not already present)
ALTER TABLE bars ADD COLUMN IF NOT EXISTS timeframe TEXT DEFAULT '1Day' NOT NULL;

-- 2. Rename the old table so we can create a partitioned replacement
ALTER TABLE bars RENAME TO bars_old;

-- 3. Create the new partitioned table layout (partitioned by the 'time' column range)
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
    
    -- In Postgres partitioning, ALL unique indexes MUST include the partition key (time).
    PRIMARY KEY (time, symbol, timeframe)
) PARTITION BY RANGE (time);

-- 4. Create historical partitions covering your 2-year lookback footprint safely
-- Since it is 2026, we absolutely need the 2024 table partition to catch early history!
CREATE TABLE bars_y2024 PARTITION OF bars FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');
CREATE TABLE bars_y2025 PARTITION OF bars FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');
CREATE TABLE bars_y2026 PARTITION OF bars FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');

-- 5. Move legacy data from the old table into the new partitioned table interface
INSERT INTO bars (time, symbol, timeframe, open, high, low, close, volume, created_at)
SELECT time, symbol, timeframe, open, high, low, close, volume, created_at FROM bars_old;

-- 6. Clean up the database by dropping the unpartitioned staging copy
DROP TABLE bars_old;