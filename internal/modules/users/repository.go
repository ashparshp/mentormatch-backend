package users

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetProfileByID(ctx context.Context, id string) (*Profile, error)
	CreateProfile(ctx context.Context, p *Profile) error
	UpdateProfile(ctx context.Context, p *Profile) error
	CompleteStudentOnboarding(ctx context.Context, profileID string, req *OnboardingRequest) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetProfileByID(ctx context.Context, id string) (*Profile, error) {
	var p Profile
	var fullName, avatarURL, phoneNumber, profileHeadline, organization, state sql.NullString
	var socialLinksJSON, skillTagsJSON, preferredLanguagesJSON string

	query := `
		SELECT
			id,
			email,
			full_name,
			avatar_url,
			phone_number,
			profile_headline,
			organization,
			COALESCE(social_links, '{}'::jsonb)::text,
			COALESCE(skill_tags, '[]'::jsonb)::text,
			COALESCE(preferred_languages, '[]'::jsonb)::text,
			state,
			role,
			has_completed_onboarding,
			created_at,
			updated_at
		FROM public.profiles
		WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.Email,
		&fullName,
		&avatarURL,
		&phoneNumber,
		&profileHeadline,
		&organization,
		&socialLinksJSON,
		&skillTagsJSON,
		&preferredLanguagesJSON,
		&state,
		&p.Role,
		&p.HasCompletedOnboarding,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	p.FullName = fullName.String
	p.AvatarURL = avatarURL.String
	p.PhoneNumber = phoneNumber.String
	p.ProfileHeadline = profileHeadline.String
	p.Organization = organization.String
	p.State = state.String

	if err := json.Unmarshal([]byte(socialLinksJSON), &p.SocialLinks); err != nil {
		return nil, fmt.Errorf("failed to parse social_links json: %w", err)
	}
	if p.SocialLinks == nil {
		p.SocialLinks = map[string]interface{}{}
	}

	if err := json.Unmarshal([]byte(skillTagsJSON), &p.SkillTags); err != nil {
		return nil, fmt.Errorf("failed to parse skill_tags json: %w", err)
	}
	if p.SkillTags == nil {
		p.SkillTags = []string{}
	}

	if err := json.Unmarshal([]byte(preferredLanguagesJSON), &p.PreferredLanguages); err != nil {
		return nil, fmt.Errorf("failed to parse preferred_languages json: %w", err)
	}
	if p.PreferredLanguages == nil {
		p.PreferredLanguages = []string{}
	}

	return &p, nil
}

func (r *repository) CreateProfile(ctx context.Context, p *Profile) error {
	query := `INSERT INTO public.profiles (id, email, full_name, avatar_url, phone_number, profile_headline, organization, social_links, skill_tags, preferred_languages, state, role, has_completed_onboarding) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := r.db.Exec(ctx, query, p.ID, p.Email, p.FullName, p.AvatarURL, p.PhoneNumber, p.ProfileHeadline, p.Organization, p.SocialLinks, p.SkillTags, p.PreferredLanguages, p.State, p.Role, p.HasCompletedOnboarding)
	if err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}
	return nil
}

func (r *repository) UpdateProfile(ctx context.Context, p *Profile) error {
	query := `UPDATE public.profiles SET full_name = $1, avatar_url = $2, phone_number = $3, profile_headline = $4, organization = $5, social_links = $6, skill_tags = $7, preferred_languages = $8, state = $9, has_completed_onboarding = $10, updated_at = timezone('utc'::text, now()) WHERE id = $11`
	_, err := r.db.Exec(ctx, query, p.FullName, p.AvatarURL, p.PhoneNumber, p.ProfileHeadline, p.Organization, p.SocialLinks, p.SkillTags, p.PreferredLanguages, p.State, p.HasCompletedOnboarding, p.ID)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}
	return nil
}

func (r *repository) CompleteStudentOnboarding(ctx context.Context, profileID string, req *OnboardingRequest) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update base profile
	profileQuery := `UPDATE public.profiles SET full_name = $1, phone_number = $2, organization = $3, skill_tags = $4, has_completed_onboarding = true, updated_at = timezone('utc'::text, now()) WHERE id = $5`
	_, err = tx.Exec(ctx, profileQuery, req.FullName, req.PhoneNumber, req.Organization, req.SkillTags, profileID)
	if err != nil {
		return fmt.Errorf("failed to update profiles table: %w", err)
	}

	// Insert into student profiles
	studentQuery := `INSERT INTO public.student_profiles (profile_id, interests, current_education, learning_goals) VALUES ($1, $2, $3, $4) ON CONFLICT (profile_id) DO UPDATE SET interests = EXCLUDED.interests, current_education = EXCLUDED.current_education, learning_goals = EXCLUDED.learning_goals`
	_, err = tx.Exec(ctx, studentQuery, profileID, req.TargetGoals, req.CurrentEducation, "") // Learning goals as empty for now or map from req
	if err != nil {
		return fmt.Errorf("failed to update student_profiles table: %w", err)
	}

	return tx.Commit(ctx)
}
