package handlers

import (
	"encoding/json"
	"net/http"
	"pullrequests/internal/domain"
	"pullrequests/internal/dtos"
	"pullrequests/internal/usecases"
)

type TeamHandler struct {
	usecase *usecases.TeamUsecase
}

func NewTeamHandler(usecase *usecases.TeamUsecase) *TeamHandler {
	return &TeamHandler{usecase: usecase}
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var req dtos.TeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request payload")
		return
	}
	defer r.Body.Close()

	result, err := h.usecase.AddTeam(r.Context(), req)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, result)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		WriteAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "team_name query parameter is required")
		return
	}

	team, err := h.usecase.GetTeam(r.Context(), teamName)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, team)
}

func (h *TeamHandler) handleDomainError(w http.ResponseWriter, err error) {
	if domainErr, ok := err.(domain.DomainError); ok {
		switch domainErr.Code() {
		case string(domain.ErrTeamExistsCode):
			WriteAPIError(w, http.StatusBadRequest, domainErr.Code(), domainErr.Message())
		case string(domain.ErrNotFoundCode):
			WriteAPIError(w, http.StatusNotFound, domainErr.Code(), domainErr.Message())
		default:
			WriteAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}
	WriteAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}
