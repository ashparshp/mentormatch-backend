package response

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, status int, success bool, data interface{}, message string, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := APIResponse{
		Success: success,
		Data:    data,
		Message: message,
		Error:   err,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func Success(w http.ResponseWriter, status int, data interface{}, message string) {
	JSON(w, status, true, data, message, "")
}

func Error(w http.ResponseWriter, status int, message string, err string) {
	JSON(w, status, false, nil, message, err)
}
