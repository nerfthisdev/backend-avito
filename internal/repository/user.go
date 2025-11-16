package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nerfthisdev/backend-avito/internal/domain"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, team_name, is_active
		FROM users
		WHERE id = $1
		`, id)

	var u domain.User
	err := row.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) SetIsActive(ctx context.Context, id string, active bool) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE users
		SET is_active = $2
		WHERE id = $1
		RETURNING id, username, team_name, is_active
		`, id, active)

	var u domain.User
	err := row.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *UserRepo) ListActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, username, team_name, is_active
           FROM users
          WHERE team_name = $1
            AND is_active = true`,
		teamName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
