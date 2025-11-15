package repository_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nerfthisdev/backend-avito/internal/domain"
	"github.com/nerfthisdev/backend-avito/internal/repository"
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

func TestTeamRepo_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t)
	defer pool.Close()

	resetDB(t, pool)

	repo := repository.NewTeamRepo(pool)

	team := domain.Team{
		Name: "backend",
		Members: []domain.User{
			{ID: "u1", Username: "Alice", IsActive: true},
			{ID: "u2", Username: "Bob", IsActive: true},
		},
	}

	if err := repo.CreateTeam(ctx, team); err != nil {
		t.Fatalf("CreateTeam: %v", err)
	}

	got, err := repo.GetByName(ctx, "backend")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}

	if got.Name != team.Name {
		t.Errorf("team name = %s, want %s", got.Name, team.Name)
	}

	if len(got.Members) != 2 {
		t.Fatalf("members len = %d, want 2", len(got.Members))
	}
}
