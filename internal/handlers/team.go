package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pullrequests/internal/domain"
	"pullrequests/internal/dtos"
	"pullrequests/internal/usecases"
	"strings"
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

	if err := h.validateTeamRequest(req); err != nil {
		WriteAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

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
		WriteAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "team_name is required")
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
		case string(domain.ErrUserExistsCode):
			WriteAPIError(w, http.StatusNotFound, domainErr.Code(), domainErr.Message())
		default:
			WriteAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}
	WriteAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}

func (h *TeamHandler) validateTeamRequest(req dtos.TeamRequest) error {
	if strings.TrimSpace(req.TeamName) == "" {
		return fmt.Errorf("team_name is required")
	}

	if len(req.Members) == 0 {
		return fmt.Errorf("team must have at least one member")
	}

	userIDs := make(map[string]bool)
	for _, member := range req.Members {
		if strings.TrimSpace(member.UserID) == "" {
			return fmt.Errorf("user_id is required for all members")
		}
		if strings.TrimSpace(member.Username) == "" {
			return fmt.Errorf("username is required for all members")
		}

		if userIDs[member.UserID] {
			return fmt.Errorf("duplicate user_id: %s", member.UserID)
		}
		userIDs[member.UserID] = true
	}

	return nil
}
