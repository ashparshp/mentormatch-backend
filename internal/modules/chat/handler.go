package chat

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

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	if bookingID == "" {
		response.Error(w, http.StatusBadRequest, "Missing booking ID", "BAD_REQUEST")
		return
	}

	senderID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed", "VALIDATION_ERROR")
		return
	}

	msg, err := h.service.SendMessage(r.Context(), bookingID, senderID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to send message", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusCreated, msg, "Message sent successfully")
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	if bookingID == "" {
		response.Error(w, http.StatusBadRequest, "Missing booking ID", "BAD_REQUEST")
		return
	}

	messages, err := h.service.GetMessages(r.Context(), bookingID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve messages", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, messages, "Messages retrieved")
}
