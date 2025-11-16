//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nerfthisdev/backend-avito/internal/domain"
	"github.com/nerfthisdev/backend-avito/internal/repository"
)

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

func seedTeamWithMembers(t *testing.T, pool *pgxpool.Pool, teamName string, members ...domain.User) {
	t.Helper()

	for i := range members {
		members[i].TeamName = teamName
	}

	repo := repository.NewTeamRepo(pool)
	if err := repo.CreateTeam(context.Background(), domain.Team{
		Name:    teamName,
		Members: members,
	}); err != nil {
		t.Fatalf("seed team %s: %v", teamName, err)
	}
}
