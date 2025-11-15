package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nerfthisdev/backend-avito/internal/domain"
)

type TeamRepo struct {
	pool *pgxpool.Pool
}

func NewTeamRepo(pool *pgxpool.Pool) *TeamRepo {
	return &TeamRepo{pool: pool}
}

func (r *TeamRepo) CreateTeam(ctx context.Context, team domain.Team) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO teams (name) VALUES ($1)`, team.Name)
	if err != nil {
		var pgErr pgx.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrTeamExists
		}
		return err
	}

	for _, m := range team.Members {
		_, err := tx.Exec(ctx, `INSERT INTO users (id, username, team_name, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE
			SET username = EXCLUDED.username,
			    team_name = EXCLUDED.team_name,
			    is_active = EXCLUDED.is_active
		`, m.ID, m.Username, team.Name, m.IsActive)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *TeamRepo) GetByName(ctx context.Context, name string) (*domain.Team, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.username, u.is_active
		FROM users u
		WHERE u.team_name = $1
		ORDER BY u.id
		`, name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var members []domain.User
	for rows.Next() {
		var u domain.User
		u.TeamName = name
		if err := rows.Scan(&u.ID, &u.Username, &u.IsActive); err != nil {
			return nil, err
		}

		members = append(members, u)

	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return nil, ErrNotFound
	}

	return &domain.Team{
		Name:    name,
		Members: members,
	}, nil
}
