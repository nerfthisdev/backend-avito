package service

import (
	"context"
	"errors"

	"github.com/nerfthisdev/backend-avito/internal/domain"
	"github.com/nerfthisdev/backend-avito/internal/repository"
)

type TeamService struct {
	teams repository.TeamRepository
}

func NewTeamService(teams repository.TeamRepository) *TeamService {
	return &TeamService{teams: teams}
}

var (
	ErrCodeTeamExists = "TEAM_EXISTS"
	ErrCodeNotFound   = "NOT_FOUND"
)

type DomainError struct {
	Code string
	Err  error
}

func (e *DomainError) Error() string { return e.Err.Error() }

func (s *TeamService) CreateTeam(ctx context.Context, team domain.Team) error {
	err := s.teams.CreateTeam(ctx, team)
	if err == nil {
		return nil
	}

	if errors.Is(err, repository.ErrTeamExists) {
		return &DomainError{Code: ErrCodeTeamExists, Err: err}
	}

	return err
}

func (s *TeamService) GetTeam(ctx context.Context, name string) (*domain.Team, error) {
	team, err := s.teams.GetByName(ctx, name)

	if err == nil {
		return team, nil
	}

	if errors.Is(err, repository.ErrNotFound) {
		return nil, &DomainError{Code: ErrCodeNotFound, Err: err}
	}
	return nil, err
}
