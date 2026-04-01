package analytics

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SearchStat struct {
	Query        string `json:"query"`
	Count        int    `json:"count"`
	AvgResults   float64 `json:"avg_results"`
	ZeroResults  int    `json:"zero_results"`
}

type Repository interface {
	StoreSearchLog(ctx context.Context, query string, resultsCount int, userID *string) error
	GetTopSearches(ctx context.Context, limit int) ([]SearchStat, error)
	GetZeroResultSearches(ctx context.Context, limit int) ([]SearchStat, error)
	GetAutocompleteSuggestions(ctx context.Context, prefix string, limit int) ([]string, error)
	GetDailySearchVolume(ctx context.Context, days int) ([]map[string]interface{}, error)
	GetTriggeredGaps(ctx context.Context, threshold int) ([]SearchStat, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) StoreSearchLog(ctx context.Context, query string, resultsCount int, userID *string) error {
	sql := `INSERT INTO public.search_logs (query, results_count, user_id) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, sql, query, resultsCount, userID)
	if err != nil {
		return fmt.Errorf("failed to store search log: %w", err)
	}
	return nil
}

func (r *repository) GetTopSearches(ctx context.Context, limit int) ([]SearchStat, error) {
	sql := `
		SELECT 
			query, 
			COUNT(*) as count, 
			AVG(results_count) as avg_results,
			COUNT(*) FILTER (WHERE results_count = 0) as zero_results
		FROM public.search_logs
		WHERE created_at > NOW() - INTERVAL '30 days'
		GROUP BY query
		ORDER BY count DESC
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, sql, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top searches: %w", err)
	}
	defer rows.Close()

	var stats []SearchStat
	for rows.Next() {
		var s SearchStat
		if err := rows.Scan(&s.Query, &s.Count, &s.AvgResults, &s.ZeroResults); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func (r *repository) GetZeroResultSearches(ctx context.Context, limit int) ([]SearchStat, error) {
	sql := `
		SELECT 
			query, 
			COUNT(*) as count
		FROM public.search_logs
		WHERE results_count = 0
		AND created_at > NOW() - INTERVAL '30 days'
		GROUP BY query
		ORDER BY count DESC
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, sql, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get zero result searches: %w", err)
	}
	defer rows.Close()

	var stats []SearchStat
	for rows.Next() {
		var s SearchStat
		if err := rows.Scan(&s.Query, &s.Count); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func (r *repository) GetAutocompleteSuggestions(ctx context.Context, prefix string, limit int) ([]string, error) {
	sql := `
		SELECT 
			query
		FROM public.search_logs
		WHERE query ILIKE $1
		AND char_length(query) >= 2
		GROUP BY query
		ORDER BY COUNT(*) DESC
		LIMIT $2
	`
	// Match queries that start with the prefix
	rows, err := r.db.Query(ctx, sql, prefix+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get autocomplete suggestions: %w", err)
	}
	defer rows.Close()

	var suggestions []string
	for rows.Next() {
		var q string
		if err := rows.Scan(&q); err != nil {
			return nil, err
		}
		suggestions = append(suggestions, q)
	}
	return suggestions, nil
}

func (r *repository) GetDailySearchVolume(ctx context.Context, days int) ([]map[string]interface{}, error) {
	sql := `
		SELECT 
			DATE(created_at) as date, 
			COUNT(*) as count
		FROM public.search_logs
		WHERE created_at > NOW() - (interval '1 day' * $1)
		GROUP BY date
		ORDER BY date ASC
	`
	rows, err := r.db.Query(ctx, sql, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily search volume: %w", err)
	}
	defer rows.Close()

	var stats []map[string]interface{}
	for rows.Next() {
		var d interface{}
		var c int
		if err := rows.Scan(&d, &c); err != nil {
			return nil, err
		}
		stats = append(stats, map[string]interface{}{
			"date":  d,
			"count": c,
		})
	}
	return stats, nil
}

func (r *repository) GetTriggeredGaps(ctx context.Context, threshold int) ([]SearchStat, error) {
	sql := `
		SELECT 
			query, 
			COUNT(*) as count
		FROM public.search_logs
		WHERE results_count = 0 
		AND created_at > NOW() - INTERVAL '24 hours'
		GROUP BY query
		HAVING COUNT(*) >= $1
		ORDER BY count DESC
	`
	rows, err := r.db.Query(ctx, sql, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to get triggered gaps: %w", err)
	}
	defer rows.Close()

	var stats []SearchStat
	for rows.Next() {
		var s SearchStat
		if err := rows.Scan(&s.Query, &s.Count); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}
