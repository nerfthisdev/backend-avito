package repository

import (
	"context"
	"errors"
	"time"

	"github.com/nerfthisdev/backend-avito/internal/domain"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrTeamExists = errors.New("team already exists")
	ErrPRExists   = errors.New("pull request already exists")
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, team domain.Team) error
	GetByName(ctx context.Context, name string) (*domain.Team, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	SetIsActive(ctx context.Context, id string, active bool) (*domain.User, error)
	ListActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr domain.PullRequest) error
	GetByID(ctx context.Context, id string) (*domain.PullRequest, error)
	SetMerged(ctx context.Context, id string, mergedAt time.Time) (*domain.PullRequest, error)
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	ListByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error)
}
