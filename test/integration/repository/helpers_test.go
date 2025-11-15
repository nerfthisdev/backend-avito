//go:build integration

package repository_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func resetDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	_, err := pool.Exec(ctx, `
        TRUNCATE TABLE pull_request_reviewers CASCADE;
        TRUNCATE TABLE pull_requests CASCADE;
        TRUNCATE TABLE users CASCADE;
        TRUNCATE TABLE teams CASCADE;
    `)
	if err != nil {
		t.Fatalf("failed to reset db: %v", err)
	}
}

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("DB_DSN_TEST")
	if dsn == "" {
		t.Fatal("test db is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	return pool
}
