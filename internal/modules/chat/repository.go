package chat

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Save(ctx context.Context, msg *Message) error
	FindByBookingID(ctx context.Context, bookingID string) ([]Message, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Save(ctx context.Context, msg *Message) error {
	query := `INSERT INTO public.chat_messages (id, booking_id, sender_id, content, is_protocol_item, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, msg.ID, msg.BookingID, msg.SenderID, msg.Content, msg.IsProtocolItem, msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}

func (r *repository) FindByBookingID(ctx context.Context, bookingID string) ([]Message, error) {
	var messages []Message
	query := `SELECT id, booking_id, sender_id, content, is_protocol_item, created_at 
			  FROM public.chat_messages WHERE booking_id = $1 ORDER BY created_at ASC`
	
	rows, err := r.db.Query(ctx, query, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m Message
		err := rows.Scan(&m.ID, &m.BookingID, &m.SenderID, &m.Content, &m.IsProtocolItem, &m.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, m)
	}

	return messages, nil
}
