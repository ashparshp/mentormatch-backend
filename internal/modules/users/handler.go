package users

import (
	"encoding/json"
	"net/http"

	"github.com/ashparshp/mentormatch-backend/internal/middleware"
	"github.com/ashparshp/mentormatch-backend/pkg/response"
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

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	profile, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get profile", "INTERNAL_ERROR")
		return
	}

	if profile == nil {
		response.Error(w, http.StatusNotFound, "Profile not found", "NOT_FOUND")
		return
	}

	response.Success(w, http.StatusOK, profile, "Profile retrieved")
}

func (h *Handler) OnboardStudent(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req OnboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed", "VALIDATION_ERROR")
		return
	}

	if err := h.service.OnboardStudent(r.Context(), userID, &req); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to onboard student", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, nil, "Student onboarding complete")
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var p Profile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	p.ID = userID // Ensure we only update the logged-in user's profile

	if err := h.service.UpdateProfile(r.Context(), &p); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update profile", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, nil, "Profile updated")
}
