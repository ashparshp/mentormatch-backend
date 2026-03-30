package mentors

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetMentorByID(ctx context.Context, id string) (*MentorProfile, error)
	GetActiveMentors(ctx context.Context, filter MentorFilter) ([]*MentorProfile, error)
	CompleteMentorOnboarding(ctx context.Context, profileID string, req *OnboardMentorRequest) error
	UpdateAvailability(ctx context.Context, profileID string, slots map[string][]string) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetMentorByID(ctx context.Context, id string) (*MentorProfile, error) {
	var m MentorProfile
	query := `
		SELECT 
			m.profile_id, m.status, m.bio, m.expertise_tags, m.verification_url, m.hourly_rate, m.availability_slots, m.verified_at, m.last_online_at,
			p.full_name, p.avatar_url, p.profile_headline, p.organization, p.skill_tags
		FROM public.mentor_profiles m
		JOIN public.profiles p ON m.profile_id = p.id
		WHERE m.profile_id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ProfileID, &m.Status, &m.Bio, &m.ExpertiseTags, &m.VerificationURL, &m.HourlyRate, &m.AvailabilitySlots, &m.VerifiedAt, &m.LastOnlineAt,
		&m.FullName, &m.AvatarURL, &m.ProfileHeadline, &m.Organization, &m.SkillTags,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get mentor profile: %w", err)
	}
	return &m, nil
}

func (r *repository) GetActiveMentors(ctx context.Context, filter MentorFilter) ([]*MentorProfile, error) {
	var mentors []*MentorProfile
	
	// Basic implementation with dynamic filters
	whereClauses := []string{"m.status = 'active'"}
	args := []interface{}{}
	argIdx := 1

	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(p.full_name ILIKE $%d OR p.profile_headline ILIKE $%d OR m.bio ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, searchTerm)
		argIdx++
	}

	if filter.Expertise != "" && filter.Expertise != "All" {
		whereClauses = append(whereClauses, fmt.Sprintf("m.expertise_tags @> $%d", argIdx))
		args = append(args, []string{filter.Expertise})
		argIdx++
	}

	if filter.MinRate > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("m.hourly_rate >= $%d", argIdx))
		args = append(args, filter.MinRate)
		argIdx++
	}

	if filter.MaxRate > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("m.hourly_rate <= $%d", argIdx))
		args = append(args, filter.MaxRate)
		argIdx++
	}

	// Calculate placeholders for Limit and Offset
	limitIdx := argIdx
	offsetIdx := argIdx + 1

	query := fmt.Sprintf(`
		SELECT 
			m.profile_id, m.status, m.bio, m.expertise_tags, m.verification_url, m.hourly_rate, m.availability_slots, m.verified_at, m.last_online_at,
			p.full_name, p.avatar_url, p.profile_headline, p.organization, p.skill_tags
		FROM public.mentor_profiles m
		JOIN public.profiles p ON m.profile_id = p.id
		WHERE %s
		LIMIT $%d OFFSET $%d
	`, strings.Join(whereClauses, " AND "), limitIdx, offsetIdx)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query mentors: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m MentorProfile
		err := rows.Scan(
			&m.ProfileID, &m.Status, &m.Bio, &m.ExpertiseTags, &m.VerificationURL, &m.HourlyRate, &m.AvailabilitySlots, &m.VerifiedAt, &m.LastOnlineAt,
			&m.FullName, &m.AvatarURL, &m.ProfileHeadline, &m.Organization, &m.SkillTags,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mentor: %w", err)
		}
		mentors = append(mentors, &m)
	}

	return mentors, nil
}

func (r *repository) CompleteMentorOnboarding(ctx context.Context, profileID string, req *OnboardMentorRequest) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update base profile (role and headline)
	profileQuery := `UPDATE public.profiles SET full_name = $1, phone_number = $2, organization = $3, skill_tags = $4, role = 'mentor', profile_headline = $5, preferred_languages = $6 WHERE id = $7`
	_, err = tx.Exec(ctx, profileQuery, req.FullName, req.PhoneNumber, req.Organization, req.SkillTags, req.ProfileHeadline, req.Languages, profileID)
	if err != nil {
		return fmt.Errorf("failed to update profiles: %w", err)
	}

	// Insert into mentor profiles
	mentorQuery := `
		INSERT INTO public.mentor_profiles (profile_id, bio, expertise_tags, verification_url, hourly_rate, availability_slots, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')
		ON CONFLICT (profile_id) DO UPDATE SET bio = EXCLUDED.bio, expertise_tags = EXCLUDED.expertise_tags, verification_url = EXCLUDED.verification_url, hourly_rate = EXCLUDED.hourly_rate, availability_slots = EXCLUDED.availability_slots
	`
	_, err = tx.Exec(ctx, mentorQuery, profileID, req.Bio, req.ExpertiseTags, req.VerificationURL, req.HourlyRate, req.AvailabilitySlots)
	if err != nil {
		return fmt.Errorf("failed to update mentor_profiles: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *repository) UpdateAvailability(ctx context.Context, profileID string, slots map[string][]string) error {
	query := `UPDATE public.mentor_profiles SET availability_slots = $1 WHERE profile_id = $2`
	_, err := r.db.Exec(ctx, query, slots, profileID)
	if err != nil {
		return fmt.Errorf("failed to update availability: %w", err)
	}
	return nil
}
