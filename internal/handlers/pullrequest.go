package handlers

import (
	"encoding/json"
	"net/http"
	"pullrequests/internal/domain"
	"pullrequests/internal/dtos"
	"pullrequests/internal/usecases"
)

type PRHandler struct {
	usecase *usecases.PRUsecase
}

func NewPRHandler(usecase *usecases.PRUsecase) *PRHandler {
	return &PRHandler{usecase: usecase}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request payload")
		return
	}
	defer r.Body.Close()

	pr, err := h.usecase.CreatePR(r.Context(), req)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, pr)
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req dtos.MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request payload")
		return
	}
	defer r.Body.Close()

	pr, err := h.usecase.MergePR(r.Context(), req)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, pr)
}

func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req dtos.ReassignPRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request payload")
		return
	}
	defer r.Body.Close()

	response, err := h.usecase.ReassignReviewer(r.Context(), req)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

func (h *PRHandler) GetUserReviewPRs(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		WriteAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "user_id query parameter is required")
		return
	}

	response, err := h.usecase.GetUserReviewPRs(r.Context(), userID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

func (h *PRHandler) handleDomainError(w http.ResponseWriter, err error) {
	if domainErr, ok := err.(domain.DomainError); ok {
		switch domainErr.Code() {
		case string(domain.ErrPRExistsCode):
			WriteAPIError(w, http.StatusConflict, "PR_EXISTS", domainErr.Message())
		case string(domain.ErrPRMergedCode):
			WriteAPIError(w, http.StatusConflict, "PR_MERGED", domainErr.Message())
		case string(domain.ErrNotAssignedCode):
			WriteAPIError(w, http.StatusConflict, "NOT_ASSIGNED", domainErr.Message())
		case string(domain.ErrNoCandidateCode):
			WriteAPIError(w, http.StatusConflict, "NO_CANDIDATE", domainErr.Message())
		case string(domain.ErrNotFoundCode):
			WriteAPIError(w, http.StatusNotFound, "NOT_FOUND", domainErr.Message())
		default:
			WriteAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}
	WriteAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}
