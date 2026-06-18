CREATE TABLE ohlcv_daily (
    id              BIGSERIAL     PRIMARY KEY,
    ticker_id       BIGINT        NOT NULL REFERENCES tickers(id),
    trade_date      DATE          NOT NULL,
    open            NUMERIC(18,6) NOT NULL,
    high            NUMERIC(18,6) NOT NULL,
    low             NUMERIC(18,6) NOT NULL,
    close           NUMERIC(18,6) NOT NULL,
    volume          NUMERIC(20,2),
    weighted_volume NUMERIC(18,6),
    transactions    BIGINT,
    adj_close       NUMERIC(18,6),
    source          VARCHAR(20)   NOT NULL,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (ticker_id, trade_date)
);

CREATE INDEX idx_ohlcv_daily_ticker_date ON ohlcv_daily (ticker_id, trade_date DESC);
