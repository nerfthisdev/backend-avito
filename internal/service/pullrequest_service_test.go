package service

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/nerfthisdev/backend-avito/internal/apperror"
	"github.com/nerfthisdev/backend-avito/internal/domain"
	"github.com/nerfthisdev/backend-avito/internal/repository"
)

func TestPullRequestService_CreateAssignsReviewers(t *testing.T) {
	t.Parallel()

	prRepo := newFakePRRepo()
	userRepo := newFakeUserRepo([]domain.User{
		{ID: "author", Username: "Author", TeamName: "backend", IsActive: true},
		{ID: "u2", Username: "B", TeamName: "backend", IsActive: true},
		{ID: "u3", Username: "C", TeamName: "backend", IsActive: true},
		{ID: "u4", Username: "D", TeamName: "backend", IsActive: false},
		{ID: "u5", Username: "E", TeamName: "payments", IsActive: true},
		{ID: "u6", Username: "F", TeamName: "backend", IsActive: true},
	})

	svc := NewPullRequestService(prRepo, userRepo)
	svc.rand = rand.New(rand.NewSource(1))

	ctx := context.Background()
	pr, err := svc.CreatePullRequest(ctx, "pr-1", "Add feature", "author")
	if err != nil {
		t.Fatalf("CreatePullRequest error: %v", err)
	}

	if len(pr.AssignedReviewers) != 2 {
		t.Fatalf("assigned len = %d, want 2", len(pr.AssignedReviewers))
	}

	for _, id := range pr.AssignedReviewers {
		if id == "author" {
			t.Fatalf("author assigned as reviewer")
		}
		if id == "u4" || id == "u5" {
			t.Fatalf("unexpected reviewer: %s", id)
		}
	}
}

func TestPullRequestService_CreateSingleReviewer(t *testing.T) {
	t.Parallel()

	prRepo := newFakePRRepo()
	userRepo := newFakeUserRepo([]domain.User{
		{ID: "author", TeamName: "backend", IsActive: true},
		{ID: "u2", TeamName: "backend", IsActive: true},
	})

	svc := NewPullRequestService(prRepo, userRepo)

	ctx := context.Background()
	pr, err := svc.CreatePullRequest(ctx, "pr-1", "Add feature", "author")
	if err != nil {
		t.Fatalf("CreatePullRequest error: %v", err)
	}

	if len(pr.AssignedReviewers) != 1 || pr.AssignedReviewers[0] != "u2" {
		t.Fatalf("unexpected reviewers: %v", pr.AssignedReviewers)
	}
}

func TestPullRequestService_CreateAuthorNotFound(t *testing.T) {
	t.Parallel()

	svc := NewPullRequestService(newFakePRRepo(), newFakeUserRepo(nil))

	_, err := svc.CreatePullRequest(context.Background(), "pr", "name", "missing")
	if code, ok := apperror.CodeOf(err); !ok || code != apperror.CodeNotFound {
		t.Fatalf("expected NOT_FOUND app error, got %v", err)
	}
}

func TestPullRequestService_ReassignReviewer(t *testing.T) {
	t.Parallel()

	prRepo := newFakePRRepo()
	prRepo.prs["pr-1"] = domain.PullRequest{
		ID:                "pr-1",
		Name:              "PR",
		AuthorID:          "author",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"rev-old", "rev-other"},
	}

	userRepo := newFakeUserRepo([]domain.User{
		{ID: "author", TeamName: "backend", IsActive: true},
		{ID: "rev-old", TeamName: "team-old", IsActive: true},
		{ID: "rev-candidate", TeamName: "team-old", IsActive: true},
		{ID: "rev-other", TeamName: "team-old", IsActive: true},
	})

	svc := NewPullRequestService(prRepo, userRepo)
	svc.rand = rand.New(rand.NewSource(2))

	ctx := context.Background()
	pr, newReviewer, err := svc.ReassignReviewer(ctx, "pr-1", "rev-old")
	if err != nil {
		t.Fatalf("ReassignReviewer error: %v", err)
	}

	if newReviewer != "rev-candidate" {
		t.Fatalf("new reviewer = %s, want rev-candidate", newReviewer)
	}
	if containsReviewer(pr.AssignedReviewers, "rev-old") {
		t.Fatalf("old reviewer still assigned: %v", pr.AssignedReviewers)
	}
	if !containsReviewer(pr.AssignedReviewers, "rev-candidate") {
		t.Fatalf("new reviewer not present: %v", pr.AssignedReviewers)
	}
}

func TestPullRequestService_ReassignNoCandidate(t *testing.T) {
	t.Parallel()

	prRepo := newFakePRRepo()
	prRepo.prs["pr-1"] = domain.PullRequest{
		ID:                "pr-1",
		Name:              "PR",
		AuthorID:          "author",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"rev-old"},
	}

	userRepo := newFakeUserRepo([]domain.User{
		{ID: "author", TeamName: "backend", IsActive: true},
		{ID: "rev-old", TeamName: "team-old", IsActive: true},
	})

	svc := NewPullRequestService(prRepo, userRepo)

	_, _, err := svc.ReassignReviewer(context.Background(), "pr-1", "rev-old")
	if !errors.Is(err, domain.ErrNoCandidate) {
		t.Fatalf("expected ErrNoCandidate, got %v", err)
	}
}

func TestPullRequestService_ReassignNotAssigned(t *testing.T) {
	t.Parallel()

	prRepo := newFakePRRepo()
	prRepo.prs["pr-1"] = domain.PullRequest{
		ID:                "pr-1",
		Name:              "PR",
		AuthorID:          "author",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"rev-other"},
	}

	userRepo := newFakeUserRepo([]domain.User{
		{ID: "author", TeamName: "backend", IsActive: true},
		{ID: "rev-old", TeamName: "team-old", IsActive: true},
		{ID: "rev-other", TeamName: "team-old", IsActive: true},
	})

	svc := NewPullRequestService(prRepo, userRepo)

	_, _, err := svc.ReassignReviewer(context.Background(), "pr-1", "rev-old")
	if !errors.Is(err, domain.ErrNotAssigned) {
		t.Fatalf("expected ErrNotAssigned, got %v", err)
	}
}

func TestPullRequestService_ReassignMerged(t *testing.T) {
	t.Parallel()

	prRepo := newFakePRRepo()
	prRepo.prs["pr-1"] = domain.PullRequest{
		ID:                "pr-1",
		Name:              "PR",
		AuthorID:          "author",
		Status:            domain.PRStatusMergerd,
		AssignedReviewers: []string{"rev-old"},
	}

	userRepo := newFakeUserRepo([]domain.User{
		{ID: "author", TeamName: "backend", IsActive: true},
		{ID: "rev-old", TeamName: "team-old", IsActive: true},
		{ID: "rev-new", TeamName: "team-old", IsActive: true},
	})

	svc := NewPullRequestService(prRepo, userRepo)

	_, _, err := svc.ReassignReviewer(context.Background(), "pr-1", "rev-old")
	if !errors.Is(err, domain.ErrPullRequestMerged) {
		t.Fatalf("expected ErrPullRequestMerged, got %v", err)
	}
}

func TestPullRequestService_MergePullRequest_Idempotent(t *testing.T) {
	t.Parallel()

	prRepo := newFakePRRepo()
	prRepo.prs["pr-1"] = domain.PullRequest{
		ID:                "pr-1",
		Name:              "PR",
		AuthorID:          "author",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"rev-old"},
	}

	svc := NewPullRequestService(prRepo, newFakeUserRepo(nil))
	svc.now = func() time.Time { return time.Unix(1000, 0).UTC() }

	ctx := context.Background()

	first, err := svc.MergePullRequest(ctx, "pr-1")
	if err != nil {
		t.Fatalf("first merge: %v", err)
	}
	if first.Status != domain.PRStatusMergerd {
		t.Fatalf("status = %s, want MERGED", first.Status)
	}
	if first.MergedAt == nil || first.MergedAt.Unix() != 1000 {
		t.Fatalf("mergedAt = %v, want 1000", first.MergedAt)
	}

	svc.now = func() time.Time { return time.Unix(2000, 0).UTC() }
	second, err := svc.MergePullRequest(ctx, "pr-1")
	if err != nil {
		t.Fatalf("second merge: %v", err)
	}
	if second.MergedAt == nil || second.MergedAt.Unix() != 1000 {
		t.Fatalf("mergedAt changed: %v", second.MergedAt)
	}
}

// --- fakes ---

type fakePRRepo struct {
	prs map[string]domain.PullRequest
}

func newFakePRRepo() *fakePRRepo {
	return &fakePRRepo{prs: make(map[string]domain.PullRequest)}
}

func (r *fakePRRepo) Create(_ context.Context, pr domain.PullRequest) error {
	if _, exists := r.prs[pr.ID]; exists {
		return repository.ErrPRExists
	}
	if pr.CreatedAt.IsZero() {
		pr.CreatedAt = time.Unix(0, 0)
	}
	r.prs[pr.ID] = clonePR(pr)
	return nil
}

func (r *fakePRRepo) GetByID(_ context.Context, id string) (*domain.PullRequest, error) {
	pr, ok := r.prs[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := clonePR(pr)
	return &cp, nil
}

func (r *fakePRRepo) SetMerged(_ context.Context, id string, mergedAt time.Time) (*domain.PullRequest, error) {
	pr, ok := r.prs[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	if pr.MergedAt == nil {
		t := mergedAt
		pr.MergedAt = &t
	}
	pr.Status = domain.PRStatusMergerd
	r.prs[id] = pr
	cp := clonePR(pr)
	return &cp, nil
}

func (r *fakePRRepo) ReplaceReviewer(_ context.Context, prID, oldReviewerID, newReviewerID string) error {
	pr, ok := r.prs[prID]
	if !ok {
		return repository.ErrNotFound
	}
	if pr.Status == domain.PRStatusMergerd {
		return domain.ErrPullRequestMerged
	}
	idx := -1
	for i, id := range pr.AssignedReviewers {
		if id == oldReviewerID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return domain.ErrNotAssigned
	}
	pr.AssignedReviewers[idx] = newReviewerID
	r.prs[prID] = pr
	return nil
}

func (r *fakePRRepo) ListByReviewer(_ context.Context, userID string) ([]domain.PullRequest, error) {
	var prs []domain.PullRequest
	for _, pr := range r.prs {
		if containsReviewer(pr.AssignedReviewers, userID) {
			prs = append(prs, clonePR(pr))
		}
	}
	return prs, nil
}

type fakeUserRepo struct {
	users map[string]domain.User
}

func newFakeUserRepo(users []domain.User) *fakeUserRepo {
	m := make(map[string]domain.User, len(users))
	for _, u := range users {
		user := u
		m[user.ID] = user
	}
	return &fakeUserRepo{users: m}
}

func (r *fakeUserRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	user := u
	return &user, nil
}

func (r *fakeUserRepo) SetIsActive(_ context.Context, id string, active bool) (*domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	u.IsActive = active
	r.users[id] = u
	user := u
	return &user, nil
}

func (r *fakeUserRepo) ListActiveByTeam(_ context.Context, teamName string) ([]domain.User, error) {
	var users []domain.User
	for _, u := range r.users {
		if u.TeamName == teamName && u.IsActive {
			user := u
			users = append(users, user)
		}
	}
	return users, nil
}

// helper to satisfy interface, not used in tests
func clonePR(pr domain.PullRequest) domain.PullRequest {
	cp := pr
	cp.AssignedReviewers = append([]string(nil), pr.AssignedReviewers...)
	if pr.MergedAt != nil {
		t := *pr.MergedAt
		cp.MergedAt = &t
	}
	return cp
}
