DROP INDEX IF EXISTS idx_tickers_figi;
ALTER TABLE tickers
  DROP COLUMN IF EXISTS is_pinned,
  DROP COLUMN IF EXISTS share_class_figi,
  DROP COLUMN IF EXISTS composite_figi;
