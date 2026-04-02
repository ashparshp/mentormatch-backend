package mentors

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ashparshp/mentormatch-backend/internal/middleware"
	"github.com/ashparshp/mentormatch-backend/internal/modules/analytics"
	"github.com/ashparshp/mentormatch-backend/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Handler struct {
	service       Service
	validate      *validator.Validate
	analyticsRepo analytics.Repository
}

func NewHandler(service Service, analyticsRepo analytics.Repository) *Handler {
	return &Handler{
		service:       service,
		validate:      validator.New(),
		analyticsRepo: analyticsRepo,
	}
}

func (h *Handler) GetMentor(w http.ResponseWriter, r *http.Request) {
	mentorID := chi.URLParam(r, "id")
	if mentorID == "" || mentorID == "undefined" {
		response.Error(w, http.StatusBadRequest, "Invalid mentor ID", "BAD_REQUEST")
		return
	}

	if _, err := uuid.Parse(mentorID); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid mentor ID", "BAD_REQUEST")
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

func (h *Handler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	mentor, err := h.service.GetMentor(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get your profile", "INTERNAL_ERROR")
		return
	}

	if mentor == nil {
		response.Success(w, http.StatusOK, nil, "No mentor profile found")
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
	if max, err := strconv.ParseFloat(query.Get("max_rate"), 64); err == nil {
		filter.MaxRate = max
	}

	filter.OrderBy = query.Get("order_by")

	mentors, err := h.service.DiscoverMentors(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to discover mentors", "INTERNAL_ERROR")
		return
	}

	// Async Search Logging
	if filter.Search != "" && h.analyticsRepo != nil {
		go func(q string, count int) {
			// We use background context for async logging to avoid cancellation if request finishes
			h.analyticsRepo.StoreSearchLog(context.Background(), q, count, nil)
		}(filter.Search, len(mentors))
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

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	profileID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req struct {
		Bio        *string          `json:"bio"`
		HourlyRate float64          `json:"hourly_rate"`
		Services   []OfferedService `json:"services"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if err := h.service.UpdateMentorProfile(r.Context(), profileID, req.Bio, req.HourlyRate, req.Services); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update mentor profile", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, nil, "Mentor profile updated")
}

func (h *Handler) GetSuggestedRate(w http.ResponseWriter, r *http.Request) {
	expertiseParam := r.URL.Query().Get("tags")
	if expertiseParam == "" {
		response.Success(w, http.StatusOK, map[string]float64{"suggested_rate": DefaultRate}, "Default rate returned")
		return
	}

	tags := strings.Split(expertiseParam, ",")
	rate := CalculateSuggestedRate(tags)

	response.Success(w, http.StatusOK, map[string]float64{"suggested_rate": rate}, "Suggested rate calculated")
}

func (h *Handler) SaveProtocolTemplates(w http.ResponseWriter, r *http.Request) {
	profileID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req SaveProtocolTemplatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if err := h.service.SaveProtocolTemplates(r.Context(), profileID, req.Templates); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to save protocol templates", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, nil, "Protocol templates saved")
}
