package reviews

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateReview(ctx context.Context, review *Review) error
	GetReviewsByMentor(ctx context.Context, mentorID string) ([]*Review, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) CreateReview(ctx context.Context, review *Review) error {
	query := `
		INSERT INTO public.reviews (booking_id, student_id, mentor_id, rating, comment)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query, 
		review.BookingID, review.StudentID, review.MentorID, review.Rating, review.Comment,
	).Scan(&review.ID, &review.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to insert review: %w", err)
	}
	return nil
}

func (r *repository) GetReviewsByMentor(ctx context.Context, mentorID string) ([]*Review, error) {
	var reviews []*Review
	query := `
		SELECT r.id, r.booking_id, r.student_id, r.mentor_id, r.rating, r.comment, r.created_at, p.full_name
		FROM public.reviews r
		JOIN public.profiles p ON r.student_id = p.id
		WHERE r.mentor_id = $1
		ORDER BY r.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, mentorID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviews: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rev Review
		err := rows.Scan(
			&rev.ID, &rev.BookingID, &rev.StudentID, &rev.MentorID, &rev.Rating, &rev.Comment, &rev.CreatedAt, &rev.StudentName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}
		reviews = append(reviews, &rev)
	}

	return reviews, nil
}
