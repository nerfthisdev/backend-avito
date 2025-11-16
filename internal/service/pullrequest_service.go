package service

import (
	"context"
	"errors"
	"math/rand"
	"slices"
	"time"

	"github.com/nerfthisdev/backend-avito/internal/apperror"
	"github.com/nerfthisdev/backend-avito/internal/domain"
	"github.com/nerfthisdev/backend-avito/internal/repository"
)

type PullRequestService struct {
	prs   repository.PullRequestRepository
	users repository.UserRepository
	rand  *rand.Rand
	now   func() time.Time
}

func NewPullRequestService(prRepo repository.PullRequestRepository, userRepo repository.UserRepository) *PullRequestService {
	return &PullRequestService{
		prs:   prRepo,
		users: userRepo,
		rand:  rand.New(rand.NewSource(time.Now().UnixNano())),
		now:   time.Now,
	}
}

func (s *PullRequestService) CreatePullRequest(ctx context.Context, id, name, authorID string) (*domain.PullRequest, error) {
	author, err := s.users.GetByID(ctx, authorID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeNotFound, err)
		}
		return nil, err
	}

	candidates, err := s.users.ListActiveByTeam(ctx, author.TeamName)
	if err != nil {
		return nil, err
	}

	reviewerIDs := s.chooseReviewers(candidates, map[string]struct{}{
		author.ID: {},
	}, 2)

	pr := domain.PullRequest{
		ID:                id,
		Name:              name,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: reviewerIDs,
	}

	if err := s.prs.Create(ctx, pr); err != nil {
		if errors.Is(err, repository.ErrPRExists) {
			return nil, domain.ErrPullRequestExists
		}
		return nil, err
	}

	return s.prs.GetByID(ctx, id)
}

func (s *PullRequestService) MergePullRequest(ctx context.Context, id string) (*domain.PullRequest, error) {
	pr, err := s.prs.SetMerged(ctx, id, s.now().UTC())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeNotFound, err)
		}
		return nil, err
	}
	return pr, nil
}

func (s *PullRequestService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	pr, err := s.prs.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", apperror.New(apperror.CodeNotFound, err)
		}
		return nil, "", err
	}

	if pr.Status == domain.PRStatusMergerd {
		return nil, "", domain.ErrPullRequestMerged
	}

	if !containsReviewer(pr.AssignedReviewers, oldReviewerID) {
		return nil, "", domain.ErrNotAssigned
	}

	reviewer, err := s.users.GetByID(ctx, oldReviewerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", apperror.New(apperror.CodeNotFound, err)
		}
		return nil, "", err
	}

	exclude := make(map[string]struct{}, len(pr.AssignedReviewers)+2)
	exclude[oldReviewerID] = struct{}{}
	exclude[pr.AuthorID] = struct{}{}
	for _, assigned := range pr.AssignedReviewers {
		exclude[assigned] = struct{}{}
	}

	candidates, err := s.users.ListActiveByTeam(ctx, reviewer.TeamName)
	if err != nil {
		return nil, "", err
	}

	next := s.chooseReviewers(candidates, exclude, 1)
	if len(next) == 0 {
		return nil, "", domain.ErrNoCandidate
	}
	newReviewerID := next[0]

	if err := s.prs.ReplaceReviewer(ctx, prID, oldReviewerID, newReviewerID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", apperror.New(apperror.CodeNotFound, err)
		}
		return nil, "", err
	}

	updated, err := s.prs.GetByID(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	return updated, newReviewerID, nil
}

func (s *PullRequestService) ListReviewerPRs(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	return s.prs.ListByReviewer(ctx, reviewerID)
}

func (s *PullRequestService) chooseReviewers(users []domain.User, exclude map[string]struct{}, limit int) []string {
	var ids []string
	for _, u := range users {
		if _, skip := exclude[u.ID]; skip {
			continue
		}
		if !u.IsActive {
			continue
		}
		ids = append(ids, u.ID)
	}

	if len(ids) <= limit {
		out := make([]string, len(ids))
		copy(out, ids)
		return out
	}

	shuffled := append([]string(nil), ids...)
	s.rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled[:limit]
}

func containsReviewer(reviewers []string, target string) bool {
	return slices.Contains(reviewers, target)
}
