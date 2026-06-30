package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TickerMigrationStore struct {
	db *pgxpool.Pool
}

func NewTickerMigrationStore(db *pgxpool.Pool) *TickerMigrationStore {
	return &TickerMigrationStore{db: db}
}

func (s *TickerMigrationStore) Create(ctx context.Context, m TickerMigration) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx, `
		INSERT INTO ticker_migrations
		       (from_ticker_id, to_ticker_id, effective_date, reason, source, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		m.FromTickerID, m.ToTickerID, m.EffectiveDate,
		string(m.Reason), string(m.Source), m.Notes,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create ticker migration: %w", err)
	}
	return id, nil
}

func (s *TickerMigrationStore) GetByFromTicker(ctx context.Context, fromTickerID int64) ([]TickerMigration, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, from_ticker_id, to_ticker_id, effective_date, reason, source,
		       notes, detected_at, created_at
		FROM ticker_migrations
		WHERE from_ticker_id = $1
		ORDER BY created_at`,
		fromTickerID,
	)
	if err != nil {
		return nil, fmt.Errorf("get migrations from ticker %d: %w", fromTickerID, err)
	}
	defer rows.Close()
	return scanMigrations(rows)
}

func (s *TickerMigrationStore) GetByToTicker(ctx context.Context, toTickerID int64) ([]TickerMigration, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, from_ticker_id, to_ticker_id, effective_date, reason, source,
		       notes, detected_at, created_at
		FROM ticker_migrations
		WHERE to_ticker_id = $1
		ORDER BY created_at`,
		toTickerID,
	)
	if err != nil {
		return nil, fmt.Errorf("get migrations to ticker %d: %w", toTickerID, err)
	}
	defer rows.Close()
	return scanMigrations(rows)
}

func scanMigrations(rows pgx.Rows) ([]TickerMigration, error) {
	var ms []TickerMigration
	for rows.Next() {
		var m TickerMigration
		var reason, source string
		if err := rows.Scan(
			&m.ID, &m.FromTickerID, &m.ToTickerID, &m.EffectiveDate,
			&reason, &source, &m.Notes, &m.DetectedAt, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan migration row: %w", err)
		}
		m.Reason = MigrationReason(reason)
		m.Source = MigrationSource(source)
		ms = append(ms, m)
	}
	return ms, rows.Err()
}
