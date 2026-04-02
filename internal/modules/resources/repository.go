package resources

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	ListResources(ctx context.Context, category string) ([]*Resource, error)
	IncrementDownloads(ctx context.Context, id string) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) ListResources(ctx context.Context, category string) ([]*Resource, error) {
	var resources []*Resource
	query := `SELECT id, title, category, COALESCE(description, ''), COALESCE(author_name, ''), downloads, rating, COALESCE(file_url, ''), created_at, updated_at FROM public.resources`
	var args []interface{}

	if category != "" && category != "all" {
		query += " WHERE category = $1"
		args = append(args, category)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query resources: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var res Resource
		err := rows.Scan(
			&res.ID, &res.Title, &res.Category, &res.Description, &res.AuthorName,
			&res.Downloads, &res.Rating, &res.FileURL, &res.CreatedAt, &res.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan resource: %w", err)
		}
		resources = append(resources, &res)
	}

	return resources, nil
}

func (r *repository) IncrementDownloads(ctx context.Context, id string) error {
	query := `UPDATE public.resources SET downloads = downloads + 1 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment downloads: %w", err)
	}
	return nil
}
