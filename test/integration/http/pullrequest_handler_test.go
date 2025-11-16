//go:build integration

package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nerfthisdev/backend-avito/internal/domain"
	httpapi "github.com/nerfthisdev/backend-avito/internal/http"
	"github.com/nerfthisdev/backend-avito/internal/http/dto"
	"github.com/nerfthisdev/backend-avito/internal/repository"
	"github.com/nerfthisdev/backend-avito/internal/service"
	"github.com/nerfthisdev/backend-avito/test/integration/testutil"
	"go.uber.org/zap"
)

func TestPullRequestHandlers_Create_Success(t *testing.T) {
	fx := newPullRequestHandlerFixture(t)

	const teamName = "backend"
	fx.insertTeam(teamName)
	fx.insertUser(domain.User{ID: "author", Username: "Alice", TeamName: teamName, IsActive: true})
	fx.insertUser(domain.User{ID: "u2", Username: "Bob", TeamName: teamName, IsActive: true})
	fx.insertUser(domain.User{ID: "u3", Username: "Charlie", TeamName: teamName, IsActive: true})

	body := mustMarshalBody(t, dto.CreatePullRequestRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Initial PR",
		AuthorID:        "author",
	})

	resp := fx.doRequest(http.MethodPost, "/pullRequest/create", body)
	if resp.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", resp.Code, http.StatusCreated, resp.Body.String())
	}

	var got struct {
		PR dto.PullRequestDTO `json:"pr"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.PR.PullRequestID != "pr-1" {
		t.Fatalf("pr id = %s, want pr-1", got.PR.PullRequestID)
	}
	if got.PR.Status != string(domain.PRStatusOpen) {
		t.Fatalf("status = %s, want OPEN", got.PR.Status)
	}
}

func TestPullRequestHandlers_Create_NotFound(t *testing.T) {
	fx := newPullRequestHandlerFixture(t)

	body := mustMarshalBody(t, dto.CreatePullRequestRequest{
		PullRequestID:   "pr-missing-author",
		PullRequestName: "Initial PR",
		AuthorID:        "missing",
	})

	resp := fx.doRequest(http.MethodPost, "/pullRequest/create", body)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d: %s", resp.Code, http.StatusNotFound, resp.Body.String())
	}
}

func TestPullRequestHandlers_Merge_Success(t *testing.T) {
	fx := newPullRequestHandlerFixture(t)

	fx.seedDefaultTeam()

	pr, err := fx.svc.CreatePullRequest(context.Background(), "pr-merge", "Feature", "author")
	if err != nil {
		t.Fatalf("CreatePullRequest: %v", err)
	}

	body := mustMarshalBody(t, dto.MergePullRequestRequest{
		PullRequestID: pr.ID,
	})

	resp := fx.doRequest(http.MethodPost, "/pullRequest/merge", body)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", resp.Code, http.StatusOK, resp.Body.String())
	}

	var got struct {
		PR dto.PullRequestDTO `json:"pr"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.PR.Status != string(domain.PRStatusMergerd) {
		t.Fatalf("status = %s, want MERGED", got.PR.Status)
	}
	if got.PR.MergedAt == nil {
		t.Fatalf("expected merged_at filled")
	}
}

func TestPullRequestHandlers_Reassign_Success(t *testing.T) {
	fx := newPullRequestHandlerFixture(t)

	fx.seedDefaultTeam()

	created, err := fx.svc.CreatePullRequest(context.Background(), "pr-reassign", "Feature", "author")
	if err != nil {
		t.Fatalf("CreatePullRequest: %v", err)
	}
	if len(created.AssignedReviewers) < 1 {
		t.Fatalf("expected reviewer assigned")
	}
	oldReviewer := created.AssignedReviewers[0]

	body := mustMarshalBody(t, dto.ReassignPullRequestRequest{
		PullRequestID: created.ID,
		OldUserID:     oldReviewer,
	})

	resp := fx.doRequest(http.MethodPost, "/pullRequest/reassign", body)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", resp.Code, http.StatusOK, resp.Body.String())
	}

	var got struct {
		PR         dto.PullRequestDTO `json:"pr"`
		ReplacedBy string             `json:"replaced_by"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ReplacedBy == "" {
		t.Fatalf("replaced_by is empty")
	}
	if got.ReplacedBy == oldReviewer {
		t.Fatalf("reassigned reviewer did not change")
	}
}

type pullRequestHandlerFixture struct {
	t    *testing.T
	pool *pgxpool.Pool
	svc  *service.PullRequestService
	mux  *http.ServeMux
}

func newPullRequestHandlerFixture(t *testing.T) *pullRequestHandlerFixture {
	t.Helper()

	pool := testutil.NewTestPool(t)
	t.Cleanup(pool.Close)
	testutil.ResetDB(t, pool)

	userRepo := repository.NewUserRepo(pool)
	prRepo := repository.NewPullRequestRepo(pool)
	svc := service.NewPullRequestService(prRepo, userRepo)
	handler := httpapi.NewPullRequestHandlers(svc, zap.NewNop())

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return &pullRequestHandlerFixture{
		t:    t,
		pool: pool,
		svc:  svc,
		mux:  mux,
	}
}

func (fx *pullRequestHandlerFixture) seedDefaultTeam() {
	const teamName = "backend"
	fx.insertTeam(teamName)
	fx.insertUser(domain.User{ID: "author", Username: "Alice", TeamName: teamName, IsActive: true})
	fx.insertUser(domain.User{ID: "u2", Username: "Bob", TeamName: teamName, IsActive: true})
	fx.insertUser(domain.User{ID: "u3", Username: "Charlie", TeamName: teamName, IsActive: true})
	fx.insertUser(domain.User{ID: "u4", Username: "Dora", TeamName: teamName, IsActive: true})
}

func (fx *pullRequestHandlerFixture) insertTeam(name string) {
	ctx := context.Background()
	if _, err := fx.pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1)`, name); err != nil {
		fx.t.Fatalf("insert team %s: %v", name, err)
	}
}

func (fx *pullRequestHandlerFixture) insertUser(user domain.User) {
	ctx := context.Background()
	_, err := fx.pool.Exec(ctx, `
		INSERT INTO users (id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
	`, user.ID, user.Username, user.TeamName, user.IsActive)
	if err != nil {
		fx.t.Fatalf("insert user %s: %v", user.ID, err)
	}
}

func (fx *pullRequestHandlerFixture) doRequest(method, target string, body []byte) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, target, reader)
	rec := httptest.NewRecorder()
	fx.mux.ServeHTTP(rec, req)
	return rec
}

func mustMarshalBody(t *testing.T, body interface{}) []byte {
	t.Helper()

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	return b
}
