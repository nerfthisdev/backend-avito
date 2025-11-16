package service

import (
	"context"
	"errors"

	"github.com/nerfthisdev/backend-avito/internal/apperror"
	"github.com/nerfthisdev/backend-avito/internal/domain"
	"github.com/nerfthisdev/backend-avito/internal/repository"
)

type TeamService struct {
	teams repository.TeamRepository
}

func NewTeamService(teams repository.TeamRepository) *TeamService {
	return &TeamService{teams: teams}
}

func (s *TeamService) CreateTeam(ctx context.Context, team domain.Team) error {
	err := s.teams.CreateTeam(ctx, team)
	if err == nil {
		return nil
	}

	if errors.Is(err, repository.ErrTeamExists) {
		return apperror.New(apperror.CodeTeamExists, err)
	}

	return err
}

func (s *TeamService) GetTeam(ctx context.Context, name string) (*domain.Team, error) {
	team, err := s.teams.GetByName(ctx, name)

	if err == nil {
		return team, nil
	}

	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.New(apperror.CodeNotFound, err)
	}
	return nil, err
}
