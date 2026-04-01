package chat

import (
	"context"
	"time"
	"github.com/google/uuid"
)

type Service interface {
	SendMessage(ctx context.Context, bookingID, senderID string, req CreateMessageRequest) (*Message, error)
	GetMessages(ctx context.Context, bookingID string) ([]Message, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) SendMessage(ctx context.Context, bookingID, senderID string, req CreateMessageRequest) (*Message, error) {
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

func (s *service) GetMessages(ctx context.Context, bookingID string) ([]Message, error) {
	return s.repo.FindByBookingID(ctx, bookingID)
}
