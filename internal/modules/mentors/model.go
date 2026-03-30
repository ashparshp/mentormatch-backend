package mentors

import (
	"time"
)

type MentorStatus string

const (
	StatusPending  MentorStatus = "pending"
	StatusActive   MentorStatus = "active"
	StatusInactive MentorStatus = "inactive"
	StatusBlocked  MentorStatus = "blocked"
)

type MentorProfile struct {
	ProfileID        string                 `json:"profile_id"`
	Status           MentorStatus           `json:"status"`
	Bio              *string                `json:"bio"`
	ExpertiseTags    []string               `json:"expertise_tags"`
	VerificationURL  *string                `json:"verification_url"`
	HourlyRate       float64                `json:"hourly_rate"`
	AvailabilitySlots map[string][]string    `json:"availability_slots"` // e.g., {"monday": ["10:00-11:00", ...]}
	VerifiedAt       *time.Time             `json:"verified_at"`
	LastOnlineAt     *time.Time             `json:"last_online_at"`
	// Joined fields from profiles
	FullName         string                 `json:"full_name,omitempty"`
	AvatarURL        *string                `json:"avatar_url,omitempty"`
	ProfileHeadline  *string                `json:"profile_headline,omitempty"`
	Organization     *string                `json:"organization,omitempty"`
	SkillTags        []string               `json:"skill_tags,omitempty"`
	AverageRating    float64                `json:"average_rating"`
	ReviewCount      int                    `json:"review_count"`
}

type OnboardMentorRequest struct {
	FullName         string              `json:"full_name" validate:"required"`
	Bio              string              `json:"bio" validate:"required"`
	ProfileHeadline  string              `json:"profile_headline" validate:"required"`
	Organization     string              `json:"organization" validate:"required"`
	VerificationURL  string              `json:"verification_url" validate:"required,url"`
	HourlyRate       float64             `json:"hourly_rate" validate:"required,gt=0"`
	ExpertiseTags    []string            `json:"expertise_tags" validate:"required,min=1"`
	SkillTags        []string            `json:"skill_tags" validate:"required,min=1"`
	AvailabilitySlots map[string][]string `json:"availability_slots" validate:"required"`
	PhoneNumber      string              `json:"phone_number" validate:"required"`
	Languages        []string            `json:"languages" validate:"required"`
}

type MentorFilter struct {
	Search    string
	Expertise string
	Tags      []string
	MinRate   float64
	MaxRate   float64
	Limit     int
	Offset    int
}
