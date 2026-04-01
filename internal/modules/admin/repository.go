package admin

import (
	"context"
	"fmt"

	"github.com/ashparshp/mentormatch-backend/internal/modules/mentors"
	"github.com/ashparshp/mentormatch-backend/internal/modules/payments"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlatformStats struct {
	TotalUsers    int     `json:"total_users"`
	TotalMentors  int     `json:"total_mentors"`
	TotalBookings int     `json:"total_bookings"`
	TotalRevenue  float64 `json:"total_revenue"`
}

type Repository interface {
	GetPlatformStats(ctx context.Context) (*PlatformStats, error)
	ListPendingMentors(ctx context.Context) ([]*mentors.MentorProfile, error)
	VerifyMentor(ctx context.Context, mentorID string, status mentors.MentorStatus) error
	ListAllPayments(ctx context.Context) ([]*payments.Payment, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetPlatformStats(ctx context.Context) (*PlatformStats, error) {
	var stats PlatformStats
	
	// 1. Total Users
	err := r.db.QueryRow(ctx, "SELECT count(*) FROM public.profiles").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// 2. Total Mentors
	err = r.db.QueryRow(ctx, "SELECT count(*) FROM public.mentor_profiles WHERE status = 'active'").Scan(&stats.TotalMentors)
	if err != nil {
		return nil, fmt.Errorf("failed to count mentors: %w", err)
	}

	// 3. Total Bookings
	err = r.db.QueryRow(ctx, "SELECT count(*) FROM public.bookings").Scan(&stats.TotalBookings)
	if err != nil {
		return nil, fmt.Errorf("failed to count bookings: %w", err)
	}

	// 4. Total Revenue (Sum of successful payments)
	err = r.db.QueryRow(ctx, "SELECT COALESCE(sum(amount), 0) FROM public.payments WHERE status = 'succeeded'").Scan(&stats.TotalRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to sum revenue: %w", err)
	}

	return &stats, nil
}

func (r *repository) ListPendingMentors(ctx context.Context) ([]*mentors.MentorProfile, error) {
	var list []*mentors.MentorProfile
	query := `
		SELECT mp.profile_id, mp.status, mp.bio, mp.expertise_tags, mp.verification_url, mp.hourly_rate,
		       p.full_name, p.organization, p.profile_headline
		FROM public.mentor_profiles mp
		JOIN public.profiles p ON mp.profile_id = p.id
		WHERE mp.status = 'pending'
		ORDER BY mp.created_at ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending mentors: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m mentors.MentorProfile
		err := rows.Scan(
			&m.ProfileID, &m.Status, &m.Bio, &m.ExpertiseTags, &m.VerificationURL, &m.HourlyRate,
			&m.FullName, &m.Organization, &m.ProfileHeadline,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending mentor: %w", err)
		}
		list = append(list, &m)
	}
	return list, nil
}

func (r *repository) VerifyMentor(ctx context.Context, mentorID string, status mentors.MentorStatus) error {
	query := `UPDATE public.mentor_profiles SET status = $1, verified_at = CASE WHEN $1 = 'active' THEN NOW() ELSE NULL END WHERE profile_id = $2`
	_, err := r.db.Exec(ctx, query, status, mentorID)
	if err != nil {
		return fmt.Errorf("failed to update mentor status: %w", err)
	}
	return nil
}

func (r *repository) ListAllPayments(ctx context.Context) ([]*payments.Payment, error) {
	var list []*payments.Payment
	query := `
		SELECT id, booking_id, student_id, amount, currency, status, provider, razorpay_order_id, razorpay_payment_id, created_at, updated_at
		FROM public.payments
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all payments: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p payments.Payment
		err := rows.Scan(
			&p.ID, &p.BookingID, &p.StudentID, &p.Amount, &p.Currency, &p.Status, &p.Provider, &p.RazorpayOrderID, &p.RazorpayPaymentID, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		list = append(list, &p)
	}
	return list, nil
}
