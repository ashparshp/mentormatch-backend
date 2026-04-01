package admin

import (
	"encoding/json"
	"net/http"

	"github.com/ashparshp/mentormatch-backend/internal/modules/mentors"
	"github.com/ashparshp/mentormatch-backend/pkg/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.GetPlatformStats(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get platform stats", "INTERNAL_ERROR")
		return
	}
	response.Success(w, http.StatusOK, stats, "Platform stats retrieved")
}

func (h *Handler) ListPendingMentors(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.ListPendingMentors(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list pending mentors", "INTERNAL_ERROR")
		return
	}
	response.Success(w, http.StatusOK, list, "Pending mentors retrieved")
}

func (h *Handler) VerifyMentor(w http.ResponseWriter, r *http.Request) {
	mentorID := chi.URLParam(r, "id")
	if mentorID == "" {
		response.Error(w, http.StatusBadRequest, "Missing mentor ID", "BAD_REQUEST")
		return
	}

	var req struct {
		Status mentors.MentorStatus `json:"status" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if err := h.repo.VerifyMentor(r.Context(), mentorID, req.Status); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to verify mentor", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, nil, "Mentor status updated")
}

func (h *Handler) ListPayments(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.ListAllPayments(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list all payments", "INTERNAL_ERROR")
		return
	}
	response.Success(w, http.StatusOK, list, "Global payment history retrieved")
}
