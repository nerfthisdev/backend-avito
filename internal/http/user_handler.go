package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/nerfthisdev/backend-avito/internal/apperror"
	"github.com/nerfthisdev/backend-avito/internal/http/dto"
	"github.com/nerfthisdev/backend-avito/internal/service"
	"go.uber.org/zap"
)

type UserHandlers struct {
	svc   *service.UserService
	prSvc *service.PullRequestService
	log   *zap.Logger
}

func NewUserHandlers(svc *service.UserService, prSvc *service.PullRequestService, log *zap.Logger) *UserHandlers {
	return &UserHandlers{svc: svc, prSvc: prSvc, log: log}
}

func (h *UserHandlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users/setIsActive", h.handleSetIsActive)
	mux.HandleFunc("GET /users/getReview", h.handleGetReview)
}

func (h *UserHandlers) handleSetIsActive(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req dto.SetIsActiveRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}

	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "user_id is required")
		return
	}

	user, err := h.svc.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		if code, ok := apperror.CodeOf(err); ok {
			if code == apperror.CodeNotFound {
				writeError(w, http.StatusNotFound, string(code), "user not found")
				return
			}
		}

		h.log.Error("failed to set user is_active", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}

	userdto := dto.UserToDTO(*user)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := struct {
		User dto.UserDTO `json:"user"`
	}{
		User: userdto,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func (h *UserHandlers) handleGetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "user_id is required")
		return
	}

	prs, err := h.prSvc.ListReviewerPRs(r.Context(), userID)
	if err != nil {
		h.log.Error("failed to list reviewer pull requests", zap.String("user_id", userID), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}

	resp := struct {
		UserID       string                    `json:"user_id"`
		PullRequests []dto.PullRequestShortDTO `json:"pull_requests"`
	}{
		UserID:       userID,
		PullRequests: dto.PullRequestsShortToDTO(prs),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
