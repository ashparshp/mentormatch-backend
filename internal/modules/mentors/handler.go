package mentors

import (
	"encoding/json"
	"net/http"
	"strconv"

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

func (h *Handler) GetMentor(w http.ResponseWriter, r *http.Request) {
	mentorID := chi.URLParam(r, "id")
	if mentorID == "" {
		response.Error(w, http.StatusBadRequest, "Missing mentor ID", "BAD_REQUEST")
		return
	}

	mentor, err := h.service.GetMentor(r.Context(), mentorID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get mentor profile", "INTERNAL_ERROR")
		return
	}

	if mentor == nil {
		response.Error(w, http.StatusNotFound, "Mentor not found", "NOT_FOUND")
		return
	}

	response.Success(w, http.StatusOK, mentor, "Mentor profile retrieved")
}

func (h *Handler) DiscoverMentors(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	
	filter := MentorFilter{
		Search:    query.Get("search"),
		Expertise: query.Get("expertise"),
		Limit:     10,
		Offset:    0,
	}

	if l, err := strconv.Atoi(query.Get("limit")); err == nil && l > 0 {
		filter.Limit = l
	}
	if o, err := strconv.Atoi(query.Get("offset")); err == nil && o >= 0 {
		filter.Offset = o
	}
	if min, err := strconv.ParseFloat(query.Get("min_rate"), 64); err == nil {
		filter.MinRate = min
	}
	if max, err := strconv.ParseFloat(query.Get("max_rate"), 64); err == nil {
		filter.MaxRate = max
	}

	mentors, err := h.service.DiscoverMentors(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to discover mentors", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, mentors, "Mentors discovered")
}

func (h *Handler) OnboardMentor(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req OnboardMentorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed", "VALIDATION_ERROR")
		return
	}

	if err := h.service.OnboardMentor(r.Context(), userID, &req); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to onboard mentor", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, nil, "Mentor onboarding submitted for approval")
}

func (h *Handler) UpdateAvailability(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req struct {
		Slots map[string][]string `json:"slots" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if err := h.service.UpdateAvailability(r.Context(), userID, req.Slots); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update availability", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, nil, "Availability updated")
}
