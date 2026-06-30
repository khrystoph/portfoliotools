package store

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SystemConfigStore struct {
	db *pgxpool.Pool
}

func NewSystemConfigStore(db *pgxpool.Pool) *SystemConfigStore {
	return &SystemConfigStore{db: db}
}

func (s *SystemConfigStore) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRow(ctx,
		`SELECT value FROM system_config WHERE key = $1`, key,
	).Scan(&value)
	if err != nil {
		return "", fmt.Errorf("get config %s: %w", key, err)
	}
	return value, nil
}

func (s *SystemConfigStore) Set(ctx context.Context, key, value string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO system_config (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`,
		key, value,
	)
	if err != nil {
		return fmt.Errorf("set config %s: %w", key, err)
	}
	return nil
}

func (s *SystemConfigStore) GetInt64(ctx context.Context, key string, fallback int64) (int64, error) {
	val, err := s.Get(ctx, key)
	if err != nil {
		if IsNotFound(err) {
			return fallback, nil
		}
		return fallback, err
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fallback, fmt.Errorf("parse config %s=%q: %w", key, val, err)
	}
	return n, nil
}
