package resources

import (
	"net/http"

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
