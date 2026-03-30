package bookings

import (
	"time"
)

type BookingStatus string

const (
	StatusPending   BookingStatus = "pending"
	StatusAccepted  BookingStatus = "accepted"
	StatusRejected  BookingStatus = "rejected"
	StatusCompleted BookingStatus = "completed"
	StatusCancelled BookingStatus = "cancelled"
)

type Booking struct {
	ID            string        `json:"id"`
	StudentID     string        `json:"student_id"`
	MentorID      string        `json:"mentor_id"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Status        BookingStatus `json:"status"`
	TotalPrice    float64       `json:"total_price"`
	MeetingLink   string        `json:"meeting_link"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`

	// Joined fields
	MentorName    string        `json:"mentor_name,omitempty"`
	MentorAvatar  string        `json:"mentor_avatar,omitempty"`
	StudentName   string        `json:"student_name,omitempty"`
	StudentAvatar string        `json:"student_avatar,omitempty"`
}

type CreateBookingRequest struct {
	MentorID  string    `json:"mentor_id" validate:"required"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
}

type UpdateBookingStatusRequest struct {
	Status      BookingStatus `json:"status" validate:"required"`
	MeetingLink string        `json:"meeting_link"`
}
