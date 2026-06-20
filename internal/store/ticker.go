package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TickerStore provides read/write access to the tickers table.
type TickerStore struct {
	db *pgxpool.Pool
}

// NewTickerStore creates a TickerStore backed by the given connection pool.
func NewTickerStore(db *pgxpool.Pool) *TickerStore {
	return &TickerStore{db: db}
}

// Upsert inserts a ticker or updates its name, primary_source, currency, and active flag
// if a ticker with the same (symbol, asset_class_id) already exists.
// Returns the ticker's ID.
func (s *TickerStore) Upsert(ctx context.Context, t Ticker) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx, `
		INSERT INTO tickers (symbol, name, asset_class_id, primary_source, currency, active)
		VALUES (
			$1, $2,
			(SELECT id FROM asset_classes WHERE name = $3),
			$4, $5, $6
		)
		ON CONFLICT (symbol, asset_class_id)
		DO UPDATE SET
			name           = EXCLUDED.name,
			primary_source = EXCLUDED.primary_source,
			currency       = EXCLUDED.currency,
			active         = EXCLUDED.active,
			updated_at     = NOW()
		RETURNING id`,
		t.Symbol, t.Name, string(t.AssetClass),
		string(t.PrimarySource), t.Currency, t.Active,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert ticker %s: %w", t.Symbol, err)
	}
	return id, nil
}

// ListActive returns all active tickers. If class is non-empty, results are
// filtered to that asset class only.
func (s *TickerStore) ListActive(ctx context.Context, class AssetClass) ([]Ticker, error) {
	rows, err := s.db.Query(ctx, `
		SELECT t.id, t.symbol, t.name, t.exchange_id, t.asset_class_id,
		       ac.name, t.primary_source, t.currency, t.active, t.created_at, t.updated_at
		FROM tickers t
		JOIN asset_classes ac ON ac.id = t.asset_class_id
		WHERE t.active = TRUE
		  AND ($1 = '' OR ac.name = $1)
		ORDER BY t.symbol`,
		string(class),
	)
	if err != nil {
		return nil, fmt.Errorf("list active tickers: %w", err)
	}
	defer rows.Close()

	var tickers []Ticker
	for rows.Next() {
		var tk Ticker
		var acName, src string
		if err := rows.Scan(
			&tk.ID, &tk.Symbol, &tk.Name, &tk.ExchangeID, &tk.AssetClassID,
			&acName, &src, &tk.Currency, &tk.Active, &tk.CreatedAt, &tk.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ticker row: %w", err)
		}
		tk.AssetClass = AssetClass(acName)
		tk.PrimarySource = DataSource(src)
		tickers = append(tickers, tk)
	}
	return tickers, rows.Err()
}

// GetBySymbol returns the ticker matching symbol and asset class.
// Returns an error wrapping pgx.ErrNoRows if not found.
func (s *TickerStore) GetBySymbol(ctx context.Context, symbol string, class AssetClass) (Ticker, error) {
	var tk Ticker
	var acName, src string
	err := s.db.QueryRow(ctx, `
		SELECT t.id, t.symbol, t.name, t.exchange_id, t.asset_class_id,
		       ac.name, t.primary_source, t.currency, t.active, t.created_at, t.updated_at
		FROM tickers t
		JOIN asset_classes ac ON ac.id = t.asset_class_id
		WHERE t.symbol = $1 AND ac.name = $2`,
		symbol, string(class),
	).Scan(
		&tk.ID, &tk.Symbol, &tk.Name, &tk.ExchangeID, &tk.AssetClassID,
		&acName, &src, &tk.Currency, &tk.Active, &tk.CreatedAt, &tk.UpdatedAt,
	)
	if err != nil {
		return Ticker{}, fmt.Errorf("get ticker %s (%s): %w", symbol, class, err)
	}
	tk.AssetClass = AssetClass(acName)
	tk.PrimarySource = DataSource(src)
	return tk, nil
}

// IsNotFound returns true when err wraps pgx.ErrNoRows, allowing callers to
// distinguish "not found" from other database errors.
func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
