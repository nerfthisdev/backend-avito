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

	httpapi "github.com/nerfthisdev/backend-avito/internal/http"
	"github.com/nerfthisdev/backend-avito/internal/http/dto"
	"github.com/nerfthisdev/backend-avito/internal/repository"
	"github.com/nerfthisdev/backend-avito/internal/service"
	"github.com/nerfthisdev/backend-avito/test/integration/testutil"
	"go.uber.org/zap"
)

func TestTeamHandler_CreateTeam_Success(t *testing.T) {
	fx := newTeamHandlerFixture(t)

	body := mustMarshal(t, dto.TeamDTO{
		TeamName: "backend",
		Members: []dto.TeamMemberDTO{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	})

	resp := fx.doRequest(http.MethodPost, "/team/add", body)
	if resp.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", resp.Code, http.StatusCreated, resp.Body.String())
	}

	var got struct {
		Team dto.TeamDTO `json:"team"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Team.TeamName != "backend" {
		t.Fatalf("team name = %s, want backend", got.Team.TeamName)
	}

	team, err := fx.repo.GetByName(context.Background(), "backend")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if len(team.Members) != 2 {
		t.Fatalf("members len = %d, want 2", len(team.Members))
	}
}

func TestTeamHandler_CreateTeam_Duplicate(t *testing.T) {
	fx := newTeamHandlerFixture(t)

	body := mustMarshal(t, dto.TeamDTO{
		TeamName: "backend",
		Members: []dto.TeamMemberDTO{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	})

	resp := fx.doRequest(http.MethodPost, "/team/add", body)
	if resp.Code != http.StatusCreated {
		t.Fatalf("initial create failed: %d %s", resp.Code, resp.Body.String())
	}

	resp = fx.doRequest(http.MethodPost, "/team/add", body)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", resp.Code, http.StatusBadRequest, resp.Body.String())
	}

	var errBody httpapi.ErrorBody
	if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if errBody.Error.Code != service.ErrCodeTeamExists {
		t.Fatalf("error code = %s, want %s", errBody.Error.Code, service.ErrCodeTeamExists)
	}
}

func TestTeamHandler_GetTeam_Success(t *testing.T) {
	fx := newTeamHandlerFixture(t)

	createBody := mustMarshal(t, dto.TeamDTO{
		TeamName: "backend",
		Members: []dto.TeamMemberDTO{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: false},
		},
	})
	resp := fx.doRequest(http.MethodPost, "/team/add", createBody)
	if resp.Code != http.StatusCreated {
		t.Fatalf("failed to create team: %d %s", resp.Code, resp.Body.String())
	}

	resp = fx.doRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", resp.Code, http.StatusOK, resp.Body.String())
	}

	var team dto.TeamDTO
	if err := json.NewDecoder(resp.Body).Decode(&team); err != nil {
		t.Fatalf("decode team: %v", err)
	}
	if team.TeamName != "backend" {
		t.Fatalf("team name = %s, want backend", team.TeamName)
	}
	if len(team.Members) != 2 {
		t.Fatalf("members len = %d, want 2", len(team.Members))
	}
}

func TestTeamHandler_GetTeam_NotFound(t *testing.T) {
	fx := newTeamHandlerFixture(t)

	resp := fx.doRequest(http.MethodGet, "/team/get?team_name=missing", nil)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d: %s", resp.Code, http.StatusNotFound, resp.Body.String())
	}

	var errBody httpapi.ErrorBody
	if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if errBody.Error.Code != "NOT_FOUND" {
		t.Fatalf("error code = %s, want NOT_FOUND", errBody.Error.Code)
	}
}

func TestTeamHandler_GetTeam_MissingParam(t *testing.T) {
	fx := newTeamHandlerFixture(t)

	resp := fx.doRequest(http.MethodGet, "/team/get", nil)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", resp.Code, http.StatusBadRequest, resp.Body.String())
	}

	var errBody httpapi.ErrorBody
	if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if errBody.Error.Code != "BAD_REQUEST" {
		t.Fatalf("error code = %s, want BAD_REQUEST", errBody.Error.Code)
	}
}

type teamHandlerFixture struct {
	repo *repository.TeamRepo
	mux  *http.ServeMux
}

func newTeamHandlerFixture(t *testing.T) *teamHandlerFixture {
	pool := testutil.NewTestPool(t)
	t.Cleanup(pool.Close)
	testutil.ResetDB(t, pool)

	repo := repository.NewTeamRepo(pool)
	svc := service.NewTeamService(repo)
	handler := httpapi.NewTeamHandler(svc, zap.NewNop())

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return &teamHandlerFixture{
		repo: repo,
		mux:  mux,
	}
}

func (fx *teamHandlerFixture) doRequest(method, target string, body []byte) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, target, reader)
	rec := httptest.NewRecorder()
	fx.mux.ServeHTTP(rec, req)
	return rec
}

func mustMarshal(t *testing.T, team dto.TeamDTO) []byte {
	t.Helper()

	b, err := json.Marshal(team)
	if err != nil {
		t.Fatalf("marshal dto: %v", err)
	}
	return b
}
