package service

import (
	"context"
	"errors"

	"github.com/nerfthisdev/backend-avito/internal/apperror"
	"github.com/nerfthisdev/backend-avito/internal/domain"
	"github.com/nerfthisdev/backend-avito/internal/repository"
)

type UserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) SetIsActive(ctx context.Context, id string, active bool) (*domain.User, error) {
	u, err := s.users.SetIsActive(ctx, id, active)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeNotFound, err)
		}
		return nil, err
	}

	return u, nil
}
