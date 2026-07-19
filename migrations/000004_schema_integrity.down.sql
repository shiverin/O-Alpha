ALTER TABLE bars DROP CONSTRAINT bars_pkey;

CREATE UNIQUE INDEX bars_unique_dataset_idx
    ON bars (time, symbol, timeframe, feed, adjustment, source);

ALTER TABLE sleeve_returns
    DROP CONSTRAINT sleeve_returns_run_id_fkey;
