package resources

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ashparshp/mentormatch-backend/pkg/response"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ListResources(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	
	resources, err := h.repo.ListResources(r.Context(), category)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list resources", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, resources, "Resources retrieved")
}

func (h *Handler) DownloadResource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Missing resource ID", "BAD_REQUEST")
		return
	}

	if err := h.repo.IncrementDownloads(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to increment download count", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, nil, "Download recorded")
}
