CREATE TABLE tickers (
    id              BIGSERIAL    PRIMARY KEY,
    symbol          VARCHAR(20)  NOT NULL,
    name            VARCHAR(200),
    exchange_id     INTEGER      REFERENCES exchanges(id),
    asset_class_id  SMALLINT     NOT NULL REFERENCES asset_classes(id),
    primary_source  VARCHAR(20)  NOT NULL DEFAULT 'alpaca',
    currency        VARCHAR(10)  NOT NULL DEFAULT 'USD',
    active          BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (symbol, asset_class_id)
);

CREATE INDEX idx_tickers_symbol       ON tickers (symbol);
CREATE INDEX idx_tickers_asset_class  ON tickers (asset_class_id);
CREATE INDEX idx_tickers_active       ON tickers (active) WHERE active = TRUE;
