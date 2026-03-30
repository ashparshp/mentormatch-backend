package resources

import (
	"time"
)

type Resource struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	AuthorName  string    `json:"author_name"`
	Downloads   int       `json:"downloads"`
	Rating      float64   `json:"rating"`
	FileURL     string    `json:"file_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
