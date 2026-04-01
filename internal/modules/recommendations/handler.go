package recommendations

import (
	"net/http"
	"strconv"

	"github.com/ashparshp/mentormatch-backend/internal/middleware"
	"github.com/ashparshp/mentormatch-backend/pkg/response"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ListRecommendedMentors(w http.ResponseWriter, r *http.Request) {
	// 1. Get Student ID from context (Auth Middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "User context missing", "UNAUTHORIZED")
		return
	}

	// 2. Parse Query Params
	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 4 // Default to 4 recommendations
	}

	// 3. Get Recommendations
	recs, err := h.repo.GetRecommendationsForStudent(r.Context(), userID, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to calculate recommendations", "INTERNAL_ERROR")
		return
	}

	// 4. Return Success
	response.Success(w, http.StatusOK, RecommendationResponse{
		Recommendations: recs,
		TotalCount:      len(recs),
	}, "Personalized recommendations retrieved")
}
