package mentors

import (
	"context"
)

type Service interface {
	GetMentor(ctx context.Context, id string) (*MentorProfile, error)
	DiscoverMentors(ctx context.Context, filter MentorFilter) ([]*MentorProfile, error)
	OnboardMentor(ctx context.Context, profileID string, req *OnboardMentorRequest) error
	UpdateAvailability(ctx context.Context, profileID string, slots map[string][]string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetMentor(ctx context.Context, id string) (*MentorProfile, error) {
	return s.repo.GetMentorByID(ctx, id)
}

func (s *service) DiscoverMentors(ctx context.Context, filter MentorFilter) ([]*MentorProfile, error) {
	if filter.Limit == 0 {
		filter.Limit = 10
	}
	return s.repo.GetActiveMentors(ctx, filter)
}

func (s *service) OnboardMentor(ctx context.Context, profileID string, req *OnboardMentorRequest) error {
	return s.repo.CompleteMentorOnboarding(ctx, profileID, req)
}

func (s *service) UpdateAvailability(ctx context.Context, profileID string, slots map[string][]string) error {
	return s.repo.UpdateAvailability(ctx, profileID, slots)
}
