package recommendations

import (
	"context"
	"fmt"
	"strings"

	"github.com/ashparshp/mentormatch-backend/internal/modules/mentors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetRecommendationsForStudent(ctx context.Context, studentID string, limit int) ([]*Recommendation, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetRecommendationsForStudent(ctx context.Context, studentID string, limit int) ([]*Recommendation, error) {
	// 1. Get Student Interests
	var studentInterests []string
	err := r.db.QueryRow(ctx, "SELECT interests FROM public.student_profiles WHERE profile_id = $1", studentID).Scan(&studentInterests)
	if err != nil {
		// If no student profile found, fallback to empty interests
		studentInterests = []string{}
	}

	// 2. Perform Match Query
	// We use the && overlap operator for initial filtering
	// We calculate common tags using a subquery/function equivalent
	query := `
		SELECT 
			m.profile_id, m.status, m.bio, m.expertise_tags, m.hourly_rate, m.average_rating, m.review_count,
			p.full_name, p.organization, p.profile_headline,
			(
				SELECT array_agg(tag)
				FROM unnest(m.expertise_tags) AS tag
				WHERE tag = ANY($1)
			) as common_tags
		FROM public.mentor_profiles m
		JOIN public.profiles p ON m.profile_id = p.id
		WHERE m.status = 'active'
		AND m.expertise_tags && $1 -- Overlap with student interests
		ORDER BY cardinality(
			ARRAY(
				SELECT tag FROM unnest(m.expertise_tags) AS tag WHERE tag = ANY($1)
			)
		) DESC, m.average_rating DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, studentInterests, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recommendations: %w", err)
	}
	defer rows.Close()

	var recs []*Recommendation
	for rows.Next() {
		var m mentors.MentorProfile
		var commonTags []string
		err := rows.Scan(
			&m.ProfileID, &m.Status, &m.Bio, &m.ExpertiseTags, &m.HourlyRate, &m.AverageRating, &m.ReviewCount,
			&m.FullName, &m.Organization, &m.ProfileHeadline,
			&commonTags,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recommendation: %w", err)
		}

		// Calculate Score (Simple overlap / max possible overlap ratio)
		score := 0.0
		if len(studentInterests) > 0 {
			score = float64(len(commonTags)) / float64(len(studentInterests))
			if score > 1.0 { score = 1.0 }
		}

		recs = append(recs, &Recommendation{
			Mentor:          &m,
			MatchScore:      score,
			CommonInterests: commonTags,
			MatchReason:     fmt.Sprintf("Matched for: %s", strings.Join(commonTags, ", ")),
		})
	}

	return recs, nil
}
