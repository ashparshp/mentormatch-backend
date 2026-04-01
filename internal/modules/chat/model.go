package chat

import (
	"time"
)

type Message struct {
	ID             string    `json:"id"`
	BookingID      string    `json:"booking_id"`
	SenderID       string    `json:"sender_id"`
	Content        string    `json:"content"`
	IsProtocolItem bool      `json:"is_protocol_item"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateMessageRequest struct {
	Content        string `json:"content" validate:"required"`
	IsProtocolItem bool   `json:"is_protocol_item"`
}
