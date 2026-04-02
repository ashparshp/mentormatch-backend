package bookings

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateBooking(ctx context.Context, b *Booking) error
	GetBookingByID(ctx context.Context, id string) (*Booking, error)
	UpdateBookingStatus(ctx context.Context, id string, status BookingStatus, meetingLink string) error
	ListBookingsByStudent(ctx context.Context, studentID string) ([]*Booking, error)
	ListBookingsByMentor(ctx context.Context, mentorID string) ([]*Booking, error)
	GetUpcomingSessions(ctx context.Context, minutes int) ([]*Booking, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) CreateBooking(ctx context.Context, b *Booking) error {
	query := `
		INSERT INTO public.bookings (student_id, mentor_id, start_time, end_time, status, total_price)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query, b.StudentID, b.MentorID, b.StartTime, b.EndTime, b.Status, b.TotalPrice).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create booking: %w", err)
	}
	return nil
}

func (r *repository) GetBookingByID(ctx context.Context, id string) (*Booking, error) {
	var b Booking
	query := `SELECT id, student_id, mentor_id, start_time, end_time, status, total_price, meeting_link, created_at, updated_at FROM public.bookings WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.StudentID, &b.MentorID, &b.StartTime, &b.EndTime, &b.Status, &b.TotalPrice, &b.MeetingLink, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	return &b, nil
}

func (r *repository) UpdateBookingStatus(ctx context.Context, id string, status BookingStatus, meetingLink string) error {
	query := `UPDATE public.bookings SET status = $1, meeting_link = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(ctx, query, status, meetingLink, id)
	if err != nil {
		return fmt.Errorf("failed to update booking status: %w", err)
	}
	return nil
}

func (r *repository) ListBookingsByStudent(ctx context.Context, studentID string) ([]*Booking, error) {
	var bookings []*Booking
	query := `
		SELECT 
			b.id, b.student_id, b.mentor_id, b.start_time, b.end_time, b.status, b.total_price, b.meeting_link, b.created_at, b.updated_at,
			p.full_name as mentor_name, COALESCE(p.avatar_url, '') as mentor_avatar
		FROM public.bookings b
		JOIN public.profiles p ON b.mentor_id = p.id
		WHERE b.student_id = $1 
		ORDER BY b.start_time DESC
	`
	rows, err := r.db.Query(ctx, query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list student bookings: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var b Booking
		err := rows.Scan(
			&b.ID, &b.StudentID, &b.MentorID, &b.StartTime, &b.EndTime, &b.Status, &b.TotalPrice, &b.MeetingLink, &b.CreatedAt, &b.UpdatedAt,
			&b.MentorName, &b.MentorAvatar,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, &b)
	}
	return bookings, nil
}

func (r *repository) ListBookingsByMentor(ctx context.Context, mentorID string) ([]*Booking, error) {
	var bookings []*Booking
	query := `
		SELECT 
			b.id, b.student_id, b.mentor_id, b.start_time, b.end_time, b.status, b.total_price, b.meeting_link, b.created_at, b.updated_at,
			p.full_name as student_name, COALESCE(p.avatar_url, '') as student_avatar
		FROM public.bookings b
		JOIN public.profiles p ON b.student_id = p.id
		WHERE b.mentor_id = $1 
		ORDER BY b.start_time DESC
	`
	rows, err := r.db.Query(ctx, query, mentorID)
	if err != nil {
		return nil, fmt.Errorf("failed to list mentor bookings: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var b Booking
		err := rows.Scan(
			&b.ID, &b.StudentID, &b.MentorID, &b.StartTime, &b.EndTime, &b.Status, &b.TotalPrice, &b.MeetingLink, &b.CreatedAt, &b.UpdatedAt,
			&b.StudentName, &b.StudentAvatar,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, &b)
	}
	return bookings, nil
}

func (r *repository) GetUpcomingSessions(ctx context.Context, minutes int) ([]*Booking, error) {
	var bookings []*Booking
	query := `
		SELECT 
			id, student_id, mentor_id, start_time, end_time, status, total_price, meeting_link, created_at, updated_at
		FROM public.bookings
		WHERE status = 'accepted' 
		AND start_time > NOW() 
		AND start_time <= NOW() + ($1 || ' minutes')::interval
	`
	rows, err := r.db.Query(ctx, query, minutes)
	if err != nil {
		return nil, fmt.Errorf("failed to query upcoming sessions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var b Booking
		err := rows.Scan(
			&b.ID, &b.StudentID, &b.MentorID, &b.StartTime, &b.EndTime, &b.Status, &b.TotalPrice, &b.MeetingLink, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan upcoming booking: %w", err)
		}
		bookings = append(bookings, &b)
	}
	return bookings, nil
}
