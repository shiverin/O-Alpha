ALTER TABLE sleeve_returns
    ADD CONSTRAINT sleeve_returns_run_id_fkey
    FOREIGN KEY (run_id)
    REFERENCES portfolio_backtest_runs(id)
    ON DELETE CASCADE;

DROP INDEX bars_unique_dataset_idx;

ALTER TABLE bars
    ADD CONSTRAINT bars_pkey
    PRIMARY KEY (time, symbol, timeframe, feed, adjustment, source);
