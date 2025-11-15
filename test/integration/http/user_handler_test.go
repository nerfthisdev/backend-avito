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

func TestUserHandlers_SetIsActive_Success(t *testing.T) {
	fx := newUserHandlerFixture(t)

	const teamName = "backend"
	fx.insertTeam(teamName)
	fx.insertUser(domain.User{
		ID:       "u1",
		Username: "Alice",
		TeamName: teamName,
		IsActive: false,
	})

	body := mustMarshalSetIsActive(t, dto.SetIsActiveRequest{
		UserID:   "u1",
		IsActive: true,
	})

	resp := fx.doRequest(http.MethodPost, "/users/setIsActive", body)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", resp.Code, http.StatusOK, resp.Body.String())
	}

	var got struct {
		User dto.UserDTO `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !got.User.IsActive {
		t.Fatalf("response user IsActive = false, want true")
	}

	user, err := fx.repo.GetByID(context.Background(), "u1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if !user.IsActive {
		t.Fatalf("stored user IsActive = false, want true")
	}
}

func TestUserHandlers_SetIsActive_NotFound(t *testing.T) {
	fx := newUserHandlerFixture(t)

	body := mustMarshalSetIsActive(t, dto.SetIsActiveRequest{
		UserID:   "missing",
		IsActive: true,
	})

	resp := fx.doRequest(http.MethodPost, "/users/setIsActive", body)
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

func TestUserHandlers_SetIsActive_InvalidJSON(t *testing.T) {
	fx := newUserHandlerFixture(t)

	resp := fx.doRequest(http.MethodPost, "/users/setIsActive", []byte("not-json"))
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

func TestUserHandlers_SetIsActive_MissingUserID(t *testing.T) {
	fx := newUserHandlerFixture(t)

	body := mustMarshalSetIsActive(t, dto.SetIsActiveRequest{
		IsActive: true,
	})

	resp := fx.doRequest(http.MethodPost, "/users/setIsActive", body)
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

type userHandlerFixture struct {
	t    *testing.T
	pool *pgxpool.Pool
	repo *repository.UserRepo
	mux  *http.ServeMux
}

func newUserHandlerFixture(t *testing.T) *userHandlerFixture {
	t.Helper()

	pool := testutil.NewTestPool(t)
	t.Cleanup(pool.Close)
	testutil.ResetDB(t, pool)

	repo := repository.NewUserRepo(pool)
	svc := service.NewUserService(*repo)
	handlers := httpapi.NewUserHandlers(svc, zap.NewNop())

	mux := http.NewServeMux()
	handlers.Register(mux)

	return &userHandlerFixture{
		t:    t,
		pool: pool,
		repo: repo,
		mux:  mux,
	}
}

func (fx *userHandlerFixture) insertTeam(name string) {
	fx.t.Helper()

	ctx := context.Background()
	if _, err := fx.pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1)`, name); err != nil {
		fx.t.Fatalf("insert team %s: %v", name, err)
	}
}

func (fx *userHandlerFixture) insertUser(user domain.User) {
	fx.t.Helper()

	ctx := context.Background()
	_, err := fx.pool.Exec(ctx, `
		INSERT INTO users (id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
	`, user.ID, user.Username, user.TeamName, user.IsActive)
	if err != nil {
		fx.t.Fatalf("insert user %s: %v", user.ID, err)
	}
}

func (fx *userHandlerFixture) doRequest(method, target string, body []byte) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, target, reader)
	rec := httptest.NewRecorder()
	fx.mux.ServeHTTP(rec, req)
	return rec
}

func mustMarshalSetIsActive(t *testing.T, req dto.SetIsActiveRequest) []byte {
	t.Helper()

	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return b
}
