package users

import (
	"time"
)

type UserRole string

const (
	RoleStudent UserRole = "student"
	RoleMentor  UserRole = "mentor"
	RoleAdmin   UserRole = "admin"
)

type Profile struct {
	ID                     string                 `json:"id"`
	Email                  string                 `json:"email" validate:"required,email"`
	FullName               string                 `json:"full_name" validate:"required"`
	AvatarURL              string                 `json:"avatar_url"`
	PhoneNumber            string                 `json:"phone_number"`
	ProfileHeadline        string                 `json:"profile_headline"`
	Organization           string                 `json:"organization"`
	SocialLinks            map[string]interface{} `json:"social_links"`
	SkillTags              []string               `json:"skill_tags"`
	PreferredLanguages     []string               `json:"preferred_languages"`
	State                  string                 `json:"state"`
	Role                   UserRole               `json:"role"`
	HasCompletedOnboarding bool                   `json:"has_completed_onboarding"`
	CreatedAt              time.Time              `json:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at"`
}

type StudentProfile struct {
	ProfileID        string   `json:"profile_id"`
	Interests        []string `json:"interests"`
	CurrentEducation string   `json:"current_education"`
	LearningGoals    string   `json:"learning_goals"`
}

type OnboardingRequest struct {
	FullName         string   `json:"full_name" validate:"required"`
	CurrentEducation string   `json:"current_education" validate:"required"`
	TargetGoals      []string `json:"target_goals" validate:"required"`
	SkillTags        []string `json:"skill_tags" validate:"required"`
	PhoneNumber      string   `json:"phone_number"`
	Organization     string   `json:"organization"`
}
