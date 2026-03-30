package reviews

import (
	"context"
	"fmt"

	"github.com/ashparshp/mentormatch-backend/internal/modules/bookings"
)

type Service interface {
	CreateReview(ctx context.Context, studentID, bookingID string, req *CreateReviewRequest) (*Review, error)
	GetMentorReviews(ctx context.Context, mentorID string) ([]*Review, error)
}

type service struct {
	repo        Repository
	bookingRepo bookings.Repository
}

func NewService(repo Repository, bookingRepo bookings.Repository) Service {
	return &service{repo: repo, bookingRepo: bookingRepo}
}

func (s *service) CreateReview(ctx context.Context, studentID, bookingID string, req *CreateReviewRequest) (*Review, error) {
	// 1. Verify booking exists and belongs to student
	booking, err := s.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	if booking.StudentID != studentID {
		return nil, fmt.Errorf("unauthorized: you can only review your own bookings")
	}

	if booking.Status != bookings.StatusCompleted {
		return nil, fmt.Errorf("bad request: you can only review completed sessions")
	}

	// 2. Create review object
	review := &Review{
		BookingID: bookingID,
		StudentID: studentID,
		MentorID:  booking.MentorID,
		Rating:    req.Rating,
		Comment:   req.Comment,
	}

	if err := s.repo.CreateReview(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

func (s *service) GetMentorReviews(ctx context.Context, mentorID string) ([]*Review, error) {
	return s.repo.GetReviewsByMentor(ctx, mentorID)
}
