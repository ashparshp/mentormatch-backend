package payments

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreatePayment(ctx context.Context, p *Payment) error
	GetPaymentByOrderID(ctx context.Context, orderID string) (*Payment, error)
	UpdatePaymentStatus(ctx context.Context, id string, status PaymentStatus, razorpayPaymentID, razorpaySignature string) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) CreatePayment(ctx context.Context, p *Payment) error {
	query := `
		INSERT INTO public.payments (booking_id, student_id, amount, currency, status, provider, provider_order_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query, p.BookingID, p.StudentID, p.Amount, p.Currency, p.Status, p.Provider, p.RazorpayOrderID).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}
	return nil
}

func (r *repository) GetPaymentByOrderID(ctx context.Context, orderID string) (*Payment, error) {
	var p Payment
	query := `
		SELECT id, booking_id, student_id, amount, currency, status, provider, COALESCE(provider_order_id, ''), COALESCE(provider_payment_id, ''), COALESCE(client_secret, ''), created_at, updated_at
		FROM public.payments
		WHERE provider_order_id = $1
	`
	err := r.db.QueryRow(ctx, query, orderID).Scan(
		&p.ID, &p.BookingID, &p.StudentID, &p.Amount, &p.Currency, &p.Status, &p.Provider, &p.RazorpayOrderID, &p.RazorpayPaymentID, &p.RazorpaySignature, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	return &p, nil
}

func (r *repository) UpdatePaymentStatus(ctx context.Context, id string, status PaymentStatus, razorpayPaymentID, razorpaySignature string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Update Payment Status
	paymentQuery := `
		UPDATE public.payments 
		SET status = $1, provider_payment_id = $2, client_secret = $3, updated_at = NOW() 
		WHERE id = $4 RETURNING booking_id
	`
	var bookingID string
	err = tx.QueryRow(ctx, paymentQuery, status, razorpayPaymentID, razorpaySignature, id).Scan(&bookingID)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// 2. If Succeeded, Update Booking Status
	if status == StatusSucceeded {
		bookingQuery := `UPDATE public.bookings SET payment_status = $1, status = 'accepted' WHERE id = $2`
		_, err = tx.Exec(ctx, bookingQuery, status, bookingID)
		if err != nil {
			return fmt.Errorf("failed to update booking after payment: %w", err)
		}
	}

	return tx.Commit(ctx)
}
