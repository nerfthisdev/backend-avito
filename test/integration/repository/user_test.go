//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nerfthisdev/backend-avito/internal/domain"
	"github.com/nerfthisdev/backend-avito/internal/repository"
)

func TestUserRepo_GetByID(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t)
	defer pool.Close()

	resetDB(t, pool)

	const teamName = "backend"
	mustInsertTeam(t, pool, teamName)
	want := domain.User{
		ID:       "u1",
		Username: "Alice",
		TeamName: teamName,
		IsActive: true,
	}
	mustInsertUser(t, pool, want)

	repo := repository.NewUserRepo(pool)

	got, err := repo.GetByID(ctx, want.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if *got != want {
		t.Fatalf("user mismatch: got %+v, want %+v", got, want)
	}
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t)
	defer pool.Close()

	resetDB(t, pool)

	repo := repository.NewUserRepo(pool)
	_, err := repo.GetByID(ctx, "missing")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserRepo_SetIsActive(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t)
	defer pool.Close()

	resetDB(t, pool)

	const teamName = "backend"
	mustInsertTeam(t, pool, teamName)
	mustInsertUser(t, pool, domain.User{
		ID:       "u1",
		Username: "Alice",
		TeamName: teamName,
		IsActive: true,
	})

	repo := repository.NewUserRepo(pool)
	updated, err := repo.SetIsActive(ctx, "u1", false)
	if err != nil {
		t.Fatalf("SetIsActive: %v", err)
	}

	if updated.IsActive {
		t.Fatalf("expected IsActive false, got true")
	}

	got, err := repo.GetByID(ctx, "u1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.IsActive {
		t.Fatalf("expected stored user IsActive false, got true")
	}
}

func TestUserRepo_SetIsActive_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t)
	defer pool.Close()

	resetDB(t, pool)

	repo := repository.NewUserRepo(pool)
	_, err := repo.SetIsActive(ctx, "missing", false)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func mustInsertTeam(t *testing.T, pool *pgxpool.Pool, name string) {
	t.Helper()

	ctx := context.Background()
	if _, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1)`, name); err != nil {
		t.Fatalf("insert team %s: %v", name, err)
	}
}

func mustInsertUser(t *testing.T, pool *pgxpool.Pool, user domain.User) {
	t.Helper()

	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
	`, user.ID, user.Username, user.TeamName, user.IsActive)
	if err != nil {
		t.Fatalf("insert user %s: %v", user.ID, err)
	}
}
