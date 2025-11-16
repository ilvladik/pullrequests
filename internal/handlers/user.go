package handlers

import (
	"encoding/json"
	"net/http"
	"pullrequests/internal/domain"
	"pullrequests/internal/dtos"
	"pullrequests/internal/usecases"
)

type UserHandler struct {
	usecase *usecases.UserUsecase
}

func NewUserHandler(usecase *usecases.UserUsecase) *UserHandler {
	return &UserHandler{usecase: usecase}
}

func (h *UserHandler) SetUserActive(w http.ResponseWriter, r *http.Request) {
	var req dtos.UserActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request payload")
		return
	}
	defer r.Body.Close()

	user, err := h.usecase.SetUserActive(r.Context(), req)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) handleDomainError(w http.ResponseWriter, err error) {
	if domainErr, ok := err.(domain.DomainError); ok {
		switch domainErr.Code() {
		case string(domain.ErrNotFoundCode):
			WriteAPIError(w, http.StatusNotFound, "NOT_FOUND", domainErr.Message())
		default:
			WriteAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}
	WriteAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}
