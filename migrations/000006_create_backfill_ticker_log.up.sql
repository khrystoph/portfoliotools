CREATE TABLE backfill_ticker_log (
    id             BIGSERIAL   PRIMARY KEY,
    run_id         BIGINT      NOT NULL REFERENCES backfill_runs(id),
    ticker_id      BIGINT      NOT NULL REFERENCES tickers(id),
    status         VARCHAR(20) NOT NULL,
    candles_stored INTEGER     NOT NULL DEFAULT 0,
    error_msg      TEXT,
    duration_ms    INTEGER,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_backfill_ticker_log_run ON backfill_ticker_log (run_id);
