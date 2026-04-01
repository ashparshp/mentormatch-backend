package analytics

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetTopSearches(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	stats, err := h.repo.GetTopSearches(r.Context(), limit)
	if err != nil {
		http.Error(w, "Failed to fetch search stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) GetZeroResultSearches(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	stats, err := h.repo.GetZeroResultSearches(r.Context(), limit)
	if err != nil {
		http.Error(w, "Failed to fetch zero result stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) GetAutocomplete(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}

	suggestions, err := h.repo.GetAutocompleteSuggestions(r.Context(), query, 5)
	if err != nil {
		http.Error(w, "Failed to fetch suggestions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

func (h *Handler) GetTrends(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 14
	}

	stats, err := h.repo.GetDailySearchVolume(r.Context(), days)
	if err != nil {
		http.Error(w, "Failed to fetch trend stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	thresholdStr := r.URL.Query().Get("threshold")
	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil || threshold <= 0 {
		threshold = 5 // Default threshold of 5 misses in 24h
	}

	alerts, err := h.repo.GetTriggeredGaps(r.Context(), threshold)
	if err != nil {
		http.Error(w, "Failed to fetch gap alerts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}
