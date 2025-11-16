package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/nerfthisdev/backend-avito/internal/apperror"
	"github.com/nerfthisdev/backend-avito/internal/http/dto"
	"github.com/nerfthisdev/backend-avito/internal/service"
	"go.uber.org/zap"
)

type TeamHandlers struct {
	svc *service.TeamService
	log *zap.Logger
}

func NewTeamHandler(svc *service.TeamService, log *zap.Logger) *TeamHandlers {
	return &TeamHandlers{svc: svc, log: log}
}

func (h *TeamHandlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /team/add", h.handleCreateTeam)
	mux.HandleFunc("GET /team/get", h.handleGetTeam)
}

func (h *TeamHandlers) handleCreateTeam(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var teamdto dto.TeamDTO

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&teamdto); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}

	team := dto.TeamFromDTO(teamdto)

	if err := h.svc.CreateTeam(r.Context(), team); err != nil {
		if code, ok := apperror.CodeOf(err); ok {
			switch code {
			case apperror.CodeTeamExists:
				writeError(w, http.StatusBadRequest, string(code), "team_name already exists")
				return
			}
		}

		h.log.Error("failed to create team", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := struct {
		Team dto.TeamDTO `json:"team"`
	}{Team: teamdto}

	_ = json.NewEncoder(w).Encode(resp)
}

func (h *TeamHandlers) handleGetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")

	if teamName == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "team name is required")
		return
	}

	team, err := h.svc.GetTeam(r.Context(), teamName)
	if err != nil {
		if code, ok := apperror.CodeOf(err); ok {
			if code == apperror.CodeNotFound {
				writeError(w, http.StatusNotFound, string(code), "team not found")
				return
			}
		}

		h.log.Error("failed to get team", zap.Error(err))

		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}

	teamdto := dto.TeamToDTO(*team)

	w.Header().Set("Context-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(teamdto)
}
