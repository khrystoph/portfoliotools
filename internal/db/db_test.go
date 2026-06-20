package db_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/khrystoph/portfoliotools/internal/db"
)

// repoRoot returns the absolute path to the repository root from this test file.
func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	// file is .../internal/db/db_test.go — root is two levels up
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func TestConnect_Success(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("portfoliotools_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pgContainer.Terminate(ctx)) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := db.Connect(ctx, connStr)
	require.NoError(t, err)
	defer pool.Close()

	var result int
	err = pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestConnect_BadDSN(t *testing.T) {
	ctx := context.Background()
	_, err := db.Connect(ctx, "postgres://nobody:wrong@localhost:1/nodb?sslmode=disable&connect_timeout=1")
	assert.Error(t, err)
}

func TestMigrate_RunsAllMigrations(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("portfoliotools_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pgContainer.Terminate(ctx)) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	migrationsPath := filepath.Join(repoRoot(), "migrations")
	require.NoError(t, db.Migrate(connStr, migrationsPath))

	// Verify all expected tables exist
	pool, err := db.Connect(ctx, connStr)
	require.NoError(t, err)
	defer pool.Close()

	tables := []string{
		"asset_classes", "exchanges", "tickers",
		"ohlcv_daily", "backfill_runs", "backfill_ticker_log",
	}
	for _, table := range tables {
		var exists bool
		err := pool.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)",
			table,
		).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "expected table %q to exist after migrations", table)
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("portfoliotools_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pgContainer.Terminate(ctx)) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Run twice — second call must not error
	migrationsPath := filepath.Join(repoRoot(), "migrations")
	require.NoError(t, db.Migrate(connStr, migrationsPath))
	require.NoError(t, db.Migrate(connStr, migrationsPath))
}
