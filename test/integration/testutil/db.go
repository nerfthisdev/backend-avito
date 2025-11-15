//go:build integration

package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ResetDB truncates all tables to give tests a clean state.
func ResetDB(t *testing.T, pool *pgxpool.Pool) {
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

// NewTestPool opens a pgx pool using DB_DSN_TEST.
func NewTestPool(t *testing.T) *pgxpool.Pool {
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
