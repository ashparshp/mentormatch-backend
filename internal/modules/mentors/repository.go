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
	UpdateMentorProfile(ctx context.Context, profileID string, bio *string, hourlyRate float64, services []OfferedService) error
	SaveProtocolTemplates(ctx context.Context, profileID string, templates []ProtocolTemplate) error
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
			m.profile_id, m.status, m.bio, m.expertise_tags, m.verification_url, m.hourly_rate, m.availability_slots, m.verified_at, m.last_online_at, m.average_rating, m.review_count, m.services, m.protocol_templates,
			p.full_name, p.avatar_url, p.profile_headline, p.organization, p.skill_tags
		FROM public.mentor_profiles m
		JOIN public.profiles p ON m.profile_id = p.id
		WHERE m.profile_id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ProfileID, &m.Status, &m.Bio, &m.ExpertiseTags, &m.VerificationURL, &m.HourlyRate, &m.AvailabilitySlots, &m.VerifiedAt, &m.LastOnlineAt, &m.AverageRating, &m.ReviewCount, &m.Services, &m.ProtocolTemplates,
		&m.FullName, &m.AvatarURL, &m.ProfileHeadline, &m.Organization, &m.SkillTags,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get mentor profile: %w", err)
	}
	m.IsVerified = m.VerifiedAt != nil
	return &m, nil
}

func (r *repository) GetActiveMentors(ctx context.Context, filter MentorFilter) ([]*MentorProfile, error) {
	var mentors []*MentorProfile

	// Basic implementation with dynamic filters
	whereClauses := []string{"m.status = 'active'"}
	args := []interface{}{}
	argIdx := 1

	// Default ordering
	orderBy := "m.average_rating DESC, m.review_count DESC"
	rankSelect := "0 as match_rank"

	if filter.Search != "" {
		searchTerm := "%" + strings.TrimSpace(filter.Search) + "%"
		searchTags := strings.Split(strings.ToLower(filter.Search), " ")

		// Optimize where clause to include tag overlap and headline matches
		whereClauses = append(whereClauses, fmt.Sprintf("(p.full_name ILIKE $%d OR p.profile_headline ILIKE $%d OR m.bio ILIKE $%d OR m.expertise_tags && $%d)", argIdx, argIdx, argIdx, argIdx+1))

		// Use a weighted rank for ordering
		rankSelect = fmt.Sprintf(`(
			CASE WHEN p.full_name ILIKE $%d THEN 10 ELSE 0 END +
			CASE WHEN m.expertise_tags && $%d THEN 8 ELSE 0 END +
			CASE WHEN p.profile_headline ILIKE $%d THEN 5 ELSE 0 END +
			CASE WHEN m.bio ILIKE $%d THEN 2 ELSE 0 END
		) as match_rank`, argIdx, argIdx+1, argIdx, argIdx)

		args = append(args, searchTerm, searchTags)
		argIdx += 2

		// If searching, prioritize match_rank
		orderBy = "match_rank DESC, m.average_rating DESC"
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

	// Override ordering if specific sort selected
	if filter.OrderBy == "price_low" {
		orderBy = "m.hourly_rate ASC"
	} else if filter.OrderBy == "price_high" {
		orderBy = "m.hourly_rate DESC"
	} else if filter.OrderBy == "newest" {
		orderBy = "m.verified_at DESC NULLS LAST"
	}

	query := fmt.Sprintf(`
		SELECT 
			m.profile_id, m.status, m.bio, m.expertise_tags, m.verification_url, m.hourly_rate, m.availability_slots, m.verified_at, m.last_online_at, m.average_rating, m.review_count, m.services,
			p.full_name, p.avatar_url, p.profile_headline, p.organization, p.skill_tags,
			%s
		FROM public.mentor_profiles m
		JOIN public.profiles p ON m.profile_id = p.id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, rankSelect, strings.Join(whereClauses, " AND "), orderBy, limitIdx, offsetIdx)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query mentors: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m MentorProfile
		var matchRank int
		err := rows.Scan(
			&m.ProfileID, &m.Status, &m.Bio, &m.ExpertiseTags, &m.VerificationURL, &m.HourlyRate, &m.AvailabilitySlots, &m.VerifiedAt, &m.LastOnlineAt, &m.AverageRating, &m.ReviewCount, &m.Services,
			&m.FullName, &m.AvatarURL, &m.ProfileHeadline, &m.Organization, &m.SkillTags,
			&matchRank,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mentor: %w", err)
		}
		m.IsVerified = m.VerifiedAt != nil
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

	// Update base profile (role, headline, and completion status)
	profileQuery := `UPDATE public.profiles SET full_name = $1, phone_number = $2, organization = $3, skill_tags = $4, role = 'mentor', profile_headline = $5, preferred_languages = $6, has_completed_onboarding = true WHERE id = $7`
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

func (r *repository) UpdateMentorProfile(ctx context.Context, profileID string, bio *string, hourlyRate float64, services []OfferedService) error {
	query := `
		UPDATE public.mentor_profiles 
		SET bio = $1, hourly_rate = $2, services = $3 
		WHERE profile_id = $4
	`
	_, err := r.db.Exec(ctx, query, bio, hourlyRate, services, profileID)
	if err != nil {
		return fmt.Errorf("failed to update mentor profile: %w", err)
	}
	return nil
}

func (r *repository) UpdateAvailability(ctx context.Context, profileID string, slots map[string][]string) error {
	query := `UPDATE public.mentor_profiles SET availability_slots = $1 WHERE profile_id = $2`
	_, err := r.db.Exec(ctx, query, slots, profileID)
	if err != nil {
		return fmt.Errorf("failed to update availability: %w", err)
	}
	return nil
}

func (r *repository) SaveProtocolTemplates(ctx context.Context, profileID string, templates []ProtocolTemplate) error {
	query := `UPDATE public.mentor_profiles SET protocol_templates = $1 WHERE profile_id = $2`
	_, err := r.db.Exec(ctx, query, templates, profileID)
	if err != nil {
		return fmt.Errorf("failed to save protocol templates: %w", err)
	}
	return nil
}
