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

	if err := h.validateCreatePRRequest(req); err != nil {
		WriteAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

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

	if err := h.validateMergePRRequest(req); err != nil {
		WriteAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

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

	if err := h.validateReassignPRRequest(req); err != nil {
		WriteAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

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
		WriteAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "user_id query parameter is required")
		return
	}

	response, err := h.usecase.GetUserReviewPRs(r.Context(), userID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

func (h *PRHandler) validateCreatePRRequest(req dtos.CreatePRRequest) error {
	if strings.TrimSpace(req.PullRequestID) == "" {
		return fmt.Errorf("pull_request_id is required")
	}
	if strings.TrimSpace(req.PullRequestName) == "" {
		return fmt.Errorf("pull_request_name is required")
	}
	if strings.TrimSpace(req.AuthorID) == "" {
		return fmt.Errorf("author_id is required")
	}
	return nil
}

func (h *PRHandler) validateMergePRRequest(req dtos.MergePRRequest) error {
	if strings.TrimSpace(req.PullRequestID) == "" {
		return fmt.Errorf("pull_request_id is required")
	}
	return nil
}

func (h *PRHandler) validateReassignPRRequest(req dtos.ReassignPRRequest) error {
	if strings.TrimSpace(req.PullRequestID) == "" {
		return fmt.Errorf("pull_request_id is required")
	}
	if strings.TrimSpace(req.OldUserID) == "" {
		return fmt.Errorf("old_user_id is required")
	}
	return nil
}

func (h *PRHandler) handleDomainError(w http.ResponseWriter, err error) {
	if domainErr, ok := err.(domain.DomainError); ok {
		switch domainErr.Code() {
		case string(domain.ErrPRExistsCode):
			WriteAPIError(w, http.StatusConflict, domainErr.Code(), domainErr.Message())
		case string(domain.ErrPRMergedCode):
			WriteAPIError(w, http.StatusConflict, domainErr.Code(), domainErr.Message())
		case string(domain.ErrNotAssignedCode):
			WriteAPIError(w, http.StatusConflict, domainErr.Code(), domainErr.Message())
		case string(domain.ErrNoCandidateCode):
			WriteAPIError(w, http.StatusConflict, domainErr.Code(), domainErr.Message())
		case string(domain.ErrNotFoundCode):
			WriteAPIError(w, http.StatusNotFound, domainErr.Code(), domainErr.Message())
		default:
			WriteAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}
	WriteAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}
