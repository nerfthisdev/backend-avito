//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nerfthisdev/backend-avito/internal/domain"
	"github.com/nerfthisdev/backend-avito/internal/repository"
	"github.com/nerfthisdev/backend-avito/test/integration/testutil"
)

func TestPullRequestRepo_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	pool := testutil.NewTestPool(t)
	defer pool.Close()

	testutil.ResetDB(t, pool)

	const teamName = "backend"
	seedTeamWithMembers(t, pool, teamName,
		domain.User{ID: "u1", Username: "Alice", IsActive: true},
		domain.User{ID: "u2", Username: "Bob", IsActive: true},
	)

	repo := repository.NewPullRequestRepo(pool)

	pr := domain.PullRequest{
		ID:                "pr-1",
		Name:              "Add feature",
		AuthorID:          "u1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u2"},
	}

	if err := repo.Create(ctx, pr); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, "pr-1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.ID != pr.ID || got.Name != pr.Name || got.AuthorID != pr.AuthorID {
		t.Fatalf("unexpected pr %+v", got)
	}

	if len(got.AssignedReviewers) != 1 || got.AssignedReviewers[0] != "u2" {
		t.Fatalf("unexpected reviewers %+v", got.AssignedReviewers)
	}

	if got.CreatedAt.IsZero() {
		t.Fatalf("created_at not set")
	}
}

func TestPullRequestRepo_CreateDuplicate(t *testing.T) {
	ctx := context.Background()
	pool := testutil.NewTestPool(t)
	defer pool.Close()

	testutil.ResetDB(t, pool)

	const teamName = "backend"
	seedTeamWithMembers(t, pool, teamName,
		domain.User{ID: "u1", Username: "Alice", IsActive: true},
	)

	repo := repository.NewPullRequestRepo(pool)

	pr := domain.PullRequest{
		ID:       "pr-dup",
		Name:     "First",
		AuthorID: "u1",
		Status:   domain.PRStatusOpen,
	}

	if err := repo.Create(ctx, pr); err != nil {
		t.Fatalf("Create: %v", err)
	}

	err := repo.Create(ctx, pr)
	if !errors.Is(err, repository.ErrPRExists) {
		t.Fatalf("expected ErrPRExists, got %v", err)
	}
}

func TestPullRequestRepo_SetMerged(t *testing.T) {
	ctx := context.Background()
	pool := testutil.NewTestPool(t)
	defer pool.Close()

	testutil.ResetDB(t, pool)

	const teamName = "backend"
	seedTeamWithMembers(t, pool, teamName,
		domain.User{ID: "u1", Username: "Alice", IsActive: true},
		domain.User{ID: "u2", Username: "Bob", IsActive: true},
	)

	repo := repository.NewPullRequestRepo(pool)

	pr := domain.PullRequest{
		ID:                "pr-merge",
		Name:              "To merge",
		AuthorID:          "u1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u2"},
	}
	if err := repo.Create(ctx, pr); err != nil {
		t.Fatalf("Create: %v", err)
	}

	mergeTime := time.Now().UTC().Truncate(time.Microsecond)

	merged, err := repo.SetMerged(ctx, pr.ID, mergeTime)
	if err != nil {
		t.Fatalf("SetMerged: %v", err)
	}
	if merged.Status != domain.PRStatusMergerd {
		t.Fatalf("status = %s, want MERGED", merged.Status)
	}
	if merged.MergedAt == nil || !merged.MergedAt.Equal(mergeTime) {
		t.Fatalf("mergedAt = %v, want %v", merged.MergedAt, mergeTime)
	}

	newTime := mergeTime.Add(time.Hour)
	mergedAgain, err := repo.SetMerged(ctx, pr.ID, newTime)
	if err != nil {
		t.Fatalf("SetMerged second: %v", err)
	}
	if mergedAgain.MergedAt == nil || !mergedAgain.MergedAt.Equal(mergeTime) {
		t.Fatalf("mergedAt changed: %v, want %v", mergedAgain.MergedAt, mergeTime)
	}
}

func TestPullRequestRepo_SetMerged_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := testutil.NewTestPool(t)
	defer pool.Close()

	testutil.ResetDB(t, pool)

	repo := repository.NewPullRequestRepo(pool)

	_, err := repo.SetMerged(ctx, "missing", time.Now())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestPullRequestRepo_ReplaceReviewer(t *testing.T) {
	ctx := context.Background()
	pool := testutil.NewTestPool(t)
	defer pool.Close()

	testutil.ResetDB(t, pool)

	const teamName = "backend"
	seedTeamWithMembers(t, pool, teamName,
		domain.User{ID: "u1", Username: "Alice", IsActive: true},
		domain.User{ID: "u2", Username: "Bob", IsActive: true},
		domain.User{ID: "u3", Username: "Charlie", IsActive: true},
		domain.User{ID: "u4", Username: "Dave", IsActive: true},
	)

	repo := repository.NewPullRequestRepo(pool)
	pr := domain.PullRequest{
		ID:                "pr-reassign",
		Name:              "Needs reassign",
		AuthorID:          "u1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u2", "u3"},
	}
	if err := repo.Create(ctx, pr); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.ReplaceReviewer(ctx, pr.ID, "u2", "u4"); err != nil {
		t.Fatalf("ReplaceReviewer: %v", err)
	}

	updated, err := repo.GetByID(ctx, pr.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	want := []string{"u3", "u4"}
	if len(updated.AssignedReviewers) != len(want) {
		t.Fatalf("reviewers len = %d, want %d", len(updated.AssignedReviewers), len(want))
	}
	for i, id := range want {
		if updated.AssignedReviewers[i] != id {
			t.Fatalf("reviewers mismatch: got %v, want %v", updated.AssignedReviewers, want)
		}
	}
}

func TestPullRequestRepo_ReplaceReviewer_Merged(t *testing.T) {
	ctx := context.Background()
	pool := testutil.NewTestPool(t)
	defer pool.Close()

	testutil.ResetDB(t, pool)

	const teamName = "backend"
	seedTeamWithMembers(t, pool, teamName,
		domain.User{ID: "u1", Username: "Alice", IsActive: true},
		domain.User{ID: "u2", Username: "Bob", IsActive: true},
		domain.User{ID: "u3", Username: "Charlie", IsActive: true},
	)

	repo := repository.NewPullRequestRepo(pool)
	pr := domain.PullRequest{
		ID:                "pr-merged",
		Name:              "Merged",
		AuthorID:          "u1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u2"},
	}
	if err := repo.Create(ctx, pr); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := repo.SetMerged(ctx, pr.ID, time.Now()); err != nil {
		t.Fatalf("SetMerged: %v", err)
	}

	err := repo.ReplaceReviewer(ctx, pr.ID, "u2", "u3")
	if !errors.Is(err, domain.ErrPullRequestMerged) {
		t.Fatalf("expected ErrPullRequestMerged, got %v", err)
	}
}

func TestPullRequestRepo_ReplaceReviewer_NotAssigned(t *testing.T) {
	ctx := context.Background()
	pool := testutil.NewTestPool(t)
	defer pool.Close()

	testutil.ResetDB(t, pool)

	const teamName = "backend"
	seedTeamWithMembers(t, pool, teamName,
		domain.User{ID: "u1", Username: "Alice", IsActive: true},
		domain.User{ID: "u2", Username: "Bob", IsActive: true},
		domain.User{ID: "u3", Username: "Charlie", IsActive: true},
	)

	repo := repository.NewPullRequestRepo(pool)
	pr := domain.PullRequest{
		ID:                "pr-noassign",
		Name:              "No assign",
		AuthorID:          "u1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u2"},
	}
	if err := repo.Create(ctx, pr); err != nil {
		t.Fatalf("Create: %v", err)
	}

	err := repo.ReplaceReviewer(ctx, pr.ID, "u3", "u2")
	if !errors.Is(err, domain.ErrNotAssigned) {
		t.Fatalf("expected ErrNotAssigned, got %v", err)
	}
}

func TestPullRequestRepo_ListByReviewer(t *testing.T) {
	ctx := context.Background()
	pool := testutil.NewTestPool(t)
	defer pool.Close()

	testutil.ResetDB(t, pool)

	const teamName = "backend"
	seedTeamWithMembers(t, pool, teamName,
		domain.User{ID: "u1", Username: "Alice", IsActive: true},
		domain.User{ID: "u2", Username: "Bob", IsActive: true},
		domain.User{ID: "u3", Username: "Charlie", IsActive: true},
	)

	repo := repository.NewPullRequestRepo(pool)

	prs := []domain.PullRequest{
		{ID: "pr-a", Name: "First", AuthorID: "u1", Status: domain.PRStatusOpen, AssignedReviewers: []string{"u2", "u3"}},
		{ID: "pr-b", Name: "Second", AuthorID: "u1", Status: domain.PRStatusOpen, AssignedReviewers: []string{"u2"}},
		{ID: "pr-c", Name: "Other", AuthorID: "u1", Status: domain.PRStatusOpen, AssignedReviewers: []string{"u3"}},
	}
	for _, pr := range prs {
		if err := repo.Create(ctx, pr); err != nil {
			t.Fatalf("Create %s: %v", pr.ID, err)
		}
	}

	list, err := repo.ListByReviewer(ctx, "u2")
	if err != nil {
		t.Fatalf("ListByReviewer: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 prs, got %d", len(list))
	}

	gotIDs := []string{list[0].ID, list[1].ID}
	wantIDs := map[string]struct{}{"pr-a": {}, "pr-b": {}}
	for _, id := range gotIDs {
		if _, ok := wantIDs[id]; !ok {
			t.Fatalf("unexpected pr id %s", id)
		}
	}
	for _, pr := range list {
		if len(pr.AssignedReviewers) == 0 {
			t.Fatalf("expected reviewers filled for %s", pr.ID)
		}
	}
}
