package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BackfillStore provides read/write access to the backfill_runs and backfill_ticker_log tables.
type BackfillStore struct {
	db *pgxpool.Pool
}

// NewBackfillStore creates a BackfillStore backed by the given connection pool.
func NewBackfillStore(db *pgxpool.Pool) *BackfillStore {
	return &BackfillStore{db: db}
}

// Create inserts a new backfill run record in "running" status and returns its ID.
func (s *BackfillStore) Create(ctx context.Context) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx,
		`INSERT INTO backfill_runs DEFAULT VALUES RETURNING id`,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create backfill run: %w", err)
	}
	return id, nil
}

// Get retrieves a backfill run by ID.
func (s *BackfillStore) Get(ctx context.Context, id int64) (BackfillRun, error) {
	var r BackfillRun
	var status string
	err := s.db.QueryRow(ctx, `
		SELECT id, started_at, completed_at, status, tickers_processed, tickers_failed, error_msg, created_at
		FROM backfill_runs
		WHERE id = $1`, id,
	).Scan(&r.ID, &r.StartedAt, &r.CompletedAt, &status,
		&r.TickersProcessed, &r.TickersFailed, &r.ErrorMsg, &r.CreatedAt)
	if err != nil {
		return BackfillRun{}, fmt.Errorf("get backfill run %d: %w", id, err)
	}
	r.Status = BackfillStatus(status)
	return r, nil
}

// Complete marks a run as completed with final ticker counts.
func (s *BackfillStore) Complete(ctx context.Context, id int64, processed, failed int) error {
	_, err := s.db.Exec(ctx, `
		UPDATE backfill_runs
		SET status = 'completed', completed_at = NOW(),
		    tickers_processed = $2, tickers_failed = $3
		WHERE id = $1`,
		id, processed, failed,
	)
	if err != nil {
		return fmt.Errorf("complete backfill run %d: %w", id, err)
	}
	return nil
}

// Fail marks a run as failed with a descriptive error message.
func (s *BackfillStore) Fail(ctx context.Context, id int64, errMsg string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE backfill_runs
		SET status = 'failed', completed_at = NOW(), error_msg = $2
		WHERE id = $1`,
		id, errMsg,
	)
	if err != nil {
		return fmt.Errorf("fail backfill run %d: %w", id, err)
	}
	return nil
}

// LogTicker records the outcome for a single ticker within a backfill run.
func (s *BackfillStore) LogTicker(ctx context.Context, l BackfillTickerLog) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO backfill_ticker_log (run_id, ticker_id, status, candles_stored, error_msg, duration_ms)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		l.RunID, l.TickerID, string(l.Status), l.CandlesStored, l.ErrorMsg, l.DurationMS,
	)
	if err != nil {
		return fmt.Errorf("log ticker %d for run %d: %w", l.TickerID, l.RunID, err)
	}
	return nil
}
