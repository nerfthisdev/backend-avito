package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/nerfthisdev/backend-avito/internal/apperror"
	"github.com/nerfthisdev/backend-avito/internal/http/dto"
	"github.com/nerfthisdev/backend-avito/internal/service"
	"go.uber.org/zap"
)

type PullRequestHandlers struct {
	svc *service.PullRequestService
	log *zap.Logger
}

func NewPullRequestHandlers(svc *service.PullRequestService, log *zap.Logger) *PullRequestHandlers {
	return &PullRequestHandlers{svc: svc, log: log}
}

func (h *PullRequestHandlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /pullRequest/create", h.handleCreate)
	mux.HandleFunc("POST /pullRequest/merge", h.handleMerge)
	mux.HandleFunc("POST /pullRequest/reassign", h.handleReassign)
}

func (h *PullRequestHandlers) handleCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req dto.CreatePullRequestRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}

	if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "pull_request_id, pull_request_name and author_id are required")
		return
	}

	pr, err := h.svc.CreatePullRequest(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		if code, ok := apperror.CodeOf(err); ok {
			switch code {
			case apperror.CodeNotFound:
				writeError(w, http.StatusNotFound, string(code), "author not found")
				return
			case apperror.CodePullReqExists:
				writeError(w, http.StatusConflict, string(code), "pull request already exists")
				return
			}
		}

		h.log.Error("failed to create pull request", zap.String("pull_request_id", req.PullRequestID), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}

	resp := struct {
		PR dto.PullRequestDTO `json:"pr"`
	}{
		PR: dto.PullRequestToDTO(*pr),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *PullRequestHandlers) handleMerge(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req dto.MergePullRequestRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}
	if req.PullRequestID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "pull_request_id is required")
		return
	}

	pr, err := h.svc.MergePullRequest(r.Context(), req.PullRequestID)
	if err != nil {
		if code, ok := apperror.CodeOf(err); ok {
			if code == apperror.CodeNotFound {
				writeError(w, http.StatusNotFound, string(code), "pull request not found")
				return
			}
		}

		h.log.Error("failed to merge pull request", zap.String("pull_request_id", req.PullRequestID), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}

	resp := struct {
		PR dto.PullRequestDTO `json:"pr"`
	}{
		PR: dto.PullRequestToDTO(*pr),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *PullRequestHandlers) handleReassign(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req dto.ReassignPullRequestRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}
	if req.PullRequestID == "" || req.OldUserID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "pull_request_id and old_user_id are required")
		return
	}

	pr, replacedBy, err := h.svc.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		if code, ok := apperror.CodeOf(err); ok {
			switch code {
			case apperror.CodeNotFound:
				writeError(w, http.StatusNotFound, string(code), "pull request or user not found")
				return
			case apperror.CodePullReqMerged:
				writeError(w, http.StatusConflict, string(code), "cannot reassign on merged PR")
				return
			case apperror.CodeNotAssigned, apperror.CodeNoCandidate:
				writeError(w, http.StatusConflict, string(code), err.Error())
				return
			}
		}
		h.log.Error("failed to reassign reviewer", zap.String("pull_request_id", req.PullRequestID), zap.String("old_user_id", req.OldUserID), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}

	resp := struct {
		PR         dto.PullRequestDTO `json:"pr"`
		ReplacedBy string             `json:"replaced_by"`
	}{
		PR:         dto.PullRequestToDTO(*pr),
		ReplacedBy: replacedBy,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
