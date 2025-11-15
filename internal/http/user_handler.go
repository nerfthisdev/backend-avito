package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/nerfthisdev/backend-avito/internal/http/dto"
	"github.com/nerfthisdev/backend-avito/internal/service"
	"go.uber.org/zap"
)

type UserHandlers struct {
	svc *service.UserService
	log *zap.Logger
}

func NewUserHandlers(svc *service.UserService, log *zap.Logger) *UserHandlers {
	return &UserHandlers{svc: svc, log: log}
}

func (h *UserHandlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /users/setIsActive", h.handleSetIsActive)
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
		if derr, ok := err.(*service.DomainError); ok {
			if derr.Code == service.ErrCodeNotFound {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
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
