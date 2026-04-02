package chat

import (
	"context"
	"fmt"
	"time"
	"github.com/ashparshp/mentormatch-backend/internal/modules/bookings"
	"github.com/google/uuid"
)

type Service interface {
	SendMessage(ctx context.Context, bookingID, senderID string, req CreateMessageRequest) (*Message, error)
	GetMessages(ctx context.Context, bookingID, userID string) ([]Message, error)
}

type service struct {
	repo        Repository
	bookingRepo bookings.Repository
}

func NewService(repo Repository, bookingRepo bookings.Repository) Service {
	return &service{repo: repo, bookingRepo: bookingRepo}
}

func (s *service) SendMessage(ctx context.Context, bookingID, senderID string, req CreateMessageRequest) (*Message, error) {
	// Verify booking ownership
	booking, err := s.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	if booking.StudentID != senderID && booking.MentorID != senderID {
		return nil, fmt.Errorf("unauthorized to send messages to this booking")
	}

	msg := &Message{
		ID:             uuid.New().String(),
		BookingID:      bookingID,
		SenderID:       senderID,
		Content:        req.Content,
		IsProtocolItem: req.IsProtocolItem,
		CreatedAt:      time.Now().UTC(),
	}

	if err := s.repo.Save(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *service) GetMessages(ctx context.Context, bookingID, userID string) ([]Message, error) {
	// Verify booking ownership
	booking, err := s.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	if booking.StudentID != userID && booking.MentorID != userID {
		return nil, fmt.Errorf("unauthorized to access these messages")
	}

	return s.repo.FindByBookingID(ctx, bookingID)
}
