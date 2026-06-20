package testutil

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	ptdb "github.com/khrystoph/portfoliotools/internal/db"
)

// repoRoot returns the absolute path to the repository root by walking up from this file.
func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	// file is .../internal/testutil/db.go — root is two levels up
	return filepath.Join(filepath.Dir(file), "..", "..")
}

// NewTestDB spins up a throwaway PostgreSQL 16 container, runs all migrations,
// and returns a connected pool. Cleanup (container termination + pool close)
// is registered on t automatically.
func NewTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("portfoliotools_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("terminate postgres container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get connection string: %v", err)
	}

	migrationsPath := filepath.Join(repoRoot(), "migrations")
	if err := ptdb.Migrate(connStr, migrationsPath); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := ptdb.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}
