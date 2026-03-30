package reviews

import (
	"encoding/json"
	"net/http"

	"github.com/ashparshp/mentormatch-backend/internal/middleware"
	"github.com/ashparshp/mentormatch-backend/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service  Service
	validate *validator.Validate
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	// Authenticated student
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	bookingID := chi.URLParam(r, "id")
	if bookingID == "" {
		response.Error(w, http.StatusBadRequest, "Missing booking ID", "BAD_REQUEST")
		return
	}

	var req CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed", "VALIDATION_ERROR")
		return
	}

	review, err := h.service.CreateReview(r.Context(), userID, bookingID, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusCreated, review, "Review submitted successfully")
}

func (h *Handler) GetMentorReviews(w http.ResponseWriter, r *http.Request) {
	mentorID := chi.URLParam(r, "id")
	if mentorID == "" {
		response.Error(w, http.StatusBadRequest, "Missing mentor ID", "BAD_REQUEST")
		return
	}

	reviews, err := h.service.GetMentorReviews(r.Context(), mentorID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get reviews", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, reviews, "Mentor reviews retrieved")
}
