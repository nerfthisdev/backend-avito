//go:build integration

package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const testDBLockKey int64 = 0x5f3759df

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

	t.Cleanup(acquireDBLock(t, dsn))

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	return pool
}

func acquireDBLock(t *testing.T, dsn string) func() {
	t.Helper()

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create lock connection: %v", err)
	}

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, testDBLockKey); err != nil {
		_ = conn.Close(ctx)
		t.Fatalf("failed to acquire test db lock: %v", err)
	}

	return func() {
		if _, err := conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, testDBLockKey); err != nil {
			t.Fatalf("failed to release test db lock: %v", err)
		}
		_ = conn.Close(context.Background())
	}
}
