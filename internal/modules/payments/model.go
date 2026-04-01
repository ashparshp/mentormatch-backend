package payments

import (
	"time"
)

type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusSucceeded PaymentStatus = "succeeded"
	StatusFailed    PaymentStatus = "failed"
	StatusRefunded  PaymentStatus = "refunded"
)

type Payment struct {
	ID                string        `json:"id"`
	BookingID         string        `json:"booking_id"`
	StudentID         string        `json:"student_id"`
	Amount            float64       `json:"amount"`
	Currency          string        `json:"currency"`
	Status            PaymentStatus `json:"status"`
	Provider          string        `json:"provider"`
	RazorpayOrderID   string        `json:"razorpay_order_id"`
	RazorpayPaymentID string        `json:"razorpay_payment_id"`
	RazorpaySignature string        `json:"razorpay_signature"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

type CreateOrderRequest struct {
	BookingID string  `json:"booking_id" validate:"required"`
	Amount    float64 `json:"amount" validate:"required,gt=0"`
	Currency  string  `json:"currency" validate:"required"`
}

type VerifyPaymentRequest struct {
	RazorpayOrderID   string `json:"razorpay_order_id" validate:"required"`
	RazorpayPaymentID string `json:"razorpay_payment_id" validate:"required"`
	RazorpaySignature string `json:"razorpay_signature" validate:"required"`
}
