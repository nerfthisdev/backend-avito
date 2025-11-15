package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nerfthisdev/backend-avito/internal/domain"
	"github.com/nerfthisdev/backend-avito/internal/repository"
)

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

func TestTeamRepo_CreateTeam_Duplicate(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t)
	defer pool.Close()

	resetDB(t, pool)

	repo := repository.NewTeamRepo(pool)
	team := domain.Team{
		Name: "backend",
		Members: []domain.User{
			{ID: "u1", Username: "Alice", IsActive: true},
		},
	}

	if err := repo.CreateTeam(ctx, team); err != nil {
		t.Fatalf("CreateTeam (first insert): %v", err)
	}

	err := repo.CreateTeam(ctx, team)
	if !errors.Is(err, repository.ErrTeamExists) {
		t.Fatalf("CreateTeam expected ErrTeamExists, got %v", err)
	}
}
