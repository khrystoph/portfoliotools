CREATE TABLE system_config (
  key         VARCHAR(100) PRIMARY KEY,
  value       TEXT         NOT NULL,
  description TEXT,
  updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO system_config (key, value, description) VALUES
  ('mcap_threshold_usd', '100000000',
   'Minimum market cap in USD for automatic ticker inclusion. Pinned tickers are exempt.');

INSERT INTO asset_classes (name)
VALUES ('bond')
ON CONFLICT (name) DO NOTHING;
