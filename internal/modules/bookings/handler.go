package bookings

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

func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	studentID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed", "VALIDATION_ERROR")
		return
	}

	booking, err := h.service.CreateBooking(r.Context(), studentID, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create booking", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusCreated, booking, "Booking created successfully")
}

func (h *Handler) GetBooking(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	if bookingID == "" {
		response.Error(w, http.StatusBadRequest, "Missing booking ID", "BAD_REQUEST")
		return
	}

	booking, err := h.service.GetBooking(r.Context(), bookingID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get booking", "INTERNAL_ERROR")
		return
	}

	if booking == nil {
		response.Error(w, http.StatusNotFound, "Booking not found", "NOT_FOUND")
		return
	}

	response.Success(w, http.StatusOK, booking, "Booking retrieved")
}

func (h *Handler) ListStudentBookings(w http.ResponseWriter, r *http.Request) {
	studentID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	bookings, err := h.service.ListStudentBookings(r.Context(), studentID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list bookings", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, bookings, "Student bookings retrieved")
}

func (h *Handler) ListMentorBookings(w http.ResponseWriter, r *http.Request) {
	mentorID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	bookings, err := h.service.ListMentorBookings(r.Context(), mentorID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list bookings", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, bookings, "Mentor bookings retrieved")
}

func (h *Handler) UpdateBookingStatus(w http.ResponseWriter, r *http.Request) {
	mentorID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	bookingID := chi.URLParam(r, "id")
	if bookingID == "" {
		response.Error(w, http.StatusBadRequest, "Missing booking ID", "BAD_REQUEST")
		return
	}

	var req UpdateBookingStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed", "VALIDATION_ERROR")
		return
	}

	if err := h.service.UpdateBookingStatus(r.Context(), mentorID, bookingID, req.Status, req.MeetingLink); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update booking status", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, nil, "Booking status updated")
}
