CREATE TABLE ticker_migrations (
  id             BIGSERIAL   PRIMARY KEY,
  from_ticker_id BIGINT      NOT NULL REFERENCES tickers(id),
  to_ticker_id   BIGINT      NOT NULL REFERENCES tickers(id),
  effective_date DATE        NOT NULL,
  reason         VARCHAR(50) NOT NULL,
  source         VARCHAR(20) NOT NULL,
  notes          TEXT,
  detected_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ticker_migrations_from ON ticker_migrations(from_ticker_id);
CREATE INDEX idx_ticker_migrations_to   ON ticker_migrations(to_ticker_id);
