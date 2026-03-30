package bookings

import (
	"context"
	"fmt"

	"github.com/ashparshp/mentormatch-backend/internal/modules/mentors"
)

type Service interface {
	CreateBooking(ctx context.Context, studentID string, req *CreateBookingRequest) (*Booking, error)
	GetBooking(ctx context.Context, id string) (*Booking, error)
	UpdateBookingStatus(ctx context.Context, mentorID, id string, status BookingStatus, meetingLink string) error
	ListStudentBookings(ctx context.Context, studentID string) ([]*Booking, error)
	ListMentorBookings(ctx context.Context, mentorID string) ([]*Booking, error)
}

type service struct {
	repo       Repository
	mentorRepo mentors.Repository
}

func NewService(repo Repository, mentorRepo mentors.Repository) Service {
	return &service{repo: repo, mentorRepo: mentorRepo}
}

func (s *service) CreateBooking(ctx context.Context, studentID string, req *CreateBookingRequest) (*Booking, error) {
	mentor, err := s.mentorRepo.GetMentorByID(ctx, req.MentorID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mentor: %w", err)
	}
	if mentor == nil {
		return nil, fmt.Errorf("mentor not found")
	}

	duration := req.EndTime.Sub(req.StartTime)
	if duration <= 0 {
		return nil, fmt.Errorf("invalid duration")
	}

	totalPrice := (duration.Hours() * mentor.HourlyRate)

	booking := &Booking{
		StudentID:  studentID,
		MentorID:   req.MentorID,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Status:     StatusPending,
		TotalPrice: totalPrice,
	}

	if err := s.repo.CreateBooking(ctx, booking); err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *service) GetBooking(ctx context.Context, id string) (*Booking, error) {
	return s.repo.GetBookingByID(ctx, id)
}

func (s *service) UpdateBookingStatus(ctx context.Context, mentorID, id string, status BookingStatus, meetingLink string) error {
	booking, err := s.repo.GetBookingByID(ctx, id)
	if err != nil {
		return err
	}
	if booking == nil {
		return fmt.Errorf("booking not found")
	}

	if booking.MentorID != mentorID {
		return fmt.Errorf("unauthorized to update this booking")
	}

	return s.repo.UpdateBookingStatus(ctx, id, status, meetingLink)
}

func (s *service) ListStudentBookings(ctx context.Context, studentID string) ([]*Booking, error) {
	return s.repo.ListBookingsByStudent(ctx, studentID)
}

func (s *service) ListMentorBookings(ctx context.Context, mentorID string) ([]*Booking, error) {
	return s.repo.ListBookingsByMentor(ctx, mentorID)
}
