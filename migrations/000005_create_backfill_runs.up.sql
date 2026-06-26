CREATE TABLE backfill_runs (
    id                BIGSERIAL   PRIMARY KEY,
    started_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at      TIMESTAMPTZ,
    status            VARCHAR(20) NOT NULL DEFAULT 'running',
    tickers_processed INTEGER     NOT NULL DEFAULT 0,
    tickers_failed    INTEGER     NOT NULL DEFAULT 0,
    error_msg         TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
