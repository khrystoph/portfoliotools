ALTER TABLE tickers
  ADD COLUMN composite_figi   VARCHAR(20) UNIQUE,
  ADD COLUMN share_class_figi VARCHAR(20),
  ADD COLUMN is_pinned        BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_tickers_figi ON tickers(composite_figi)
  WHERE composite_figi IS NOT NULL;
