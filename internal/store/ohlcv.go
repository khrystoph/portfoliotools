package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// OHLCVStore provides read/write access to the ohlcv_daily table.
type OHLCVStore struct {
	db *pgxpool.Pool
}

// NewOHLCVStore creates an OHLCVStore backed by the given connection pool.
func NewOHLCVStore(db *pgxpool.Pool) *OHLCVStore {
	return &OHLCVStore{db: db}
}

const upsertOHLCVSQL = `
	INSERT INTO ohlcv_daily
		(ticker_id, trade_date, open, high, low, close, volume, weighted_volume, transactions, adj_close, source)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	ON CONFLICT (ticker_id, trade_date)
	DO UPDATE SET
		open            = EXCLUDED.open,
		high            = EXCLUDED.high,
		low             = EXCLUDED.low,
		close           = EXCLUDED.close,
		volume          = EXCLUDED.volume,
		weighted_volume = EXCLUDED.weighted_volume,
		transactions    = EXCLUDED.transactions,
		adj_close       = EXCLUDED.adj_close,
		source          = EXCLUDED.source,
		updated_at      = NOW()`

// Upsert inserts a daily candle or overwrites it if (ticker_id, trade_date) already exists.
func (s *OHLCVStore) Upsert(ctx context.Context, c OHLCVDaily) error {
	_, err := s.db.Exec(ctx, upsertOHLCVSQL,
		c.TickerID, c.TradeDate, c.Open, c.High, c.Low, c.Close,
		c.Volume, c.WeightedVolume, c.Transactions, c.AdjClose,
		string(c.Source),
	)
	if err != nil {
		return fmt.Errorf("upsert ohlcv ticker %d on %s: %w", c.TickerID, c.TradeDate.Format("2006-01-02"), err)
	}
	return nil
}

// UpsertBatch upserts all candles in a single transaction.
// If any candle fails, the entire batch is rolled back.
func (s *OHLCVStore) UpsertBatch(ctx context.Context, candles []OHLCVDaily) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin upsert batch transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, c := range candles {
		if _, err := tx.Exec(ctx, upsertOHLCVSQL,
			c.TickerID, c.TradeDate, c.Open, c.High, c.Low, c.Close,
			c.Volume, c.WeightedVolume, c.Transactions, c.AdjClose,
			string(c.Source),
		); err != nil {
			return fmt.Errorf("upsert batch — ticker %d on %s: %w",
				c.TickerID, c.TradeDate.Format("2006-01-02"), err)
		}
	}
	return tx.Commit(ctx)
}

// GetRange returns the most recent limit candles for tickerID, ordered newest first.
func (s *OHLCVStore) GetRange(ctx context.Context, tickerID int64, limit int) ([]OHLCVDaily, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, ticker_id, trade_date, open, high, low, close,
		       volume, weighted_volume, transactions, adj_close, source,
		       created_at, updated_at
		FROM ohlcv_daily
		WHERE ticker_id = $1
		ORDER BY trade_date DESC
		LIMIT $2`,
		tickerID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get ohlcv range for ticker %d: %w", tickerID, err)
	}
	defer rows.Close()

	var candles []OHLCVDaily
	for rows.Next() {
		var c OHLCVDaily
		var src string
		if err := rows.Scan(
			&c.ID, &c.TickerID, &c.TradeDate, &c.Open, &c.High, &c.Low, &c.Close,
			&c.Volume, &c.WeightedVolume, &c.Transactions, &c.AdjClose, &src,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ohlcv row: %w", err)
		}
		c.Source = DataSource(src)
		candles = append(candles, c)
	}
	return candles, rows.Err()
}
