package users

import (
	"context"
)

type Service interface {
	GetProfile(ctx context.Context, id string) (*Profile, error)
	OnboardStudent(ctx context.Context, profileID string, req *OnboardingRequest) error
	UpdateProfile(ctx context.Context, p *Profile) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetProfile(ctx context.Context, id string) (*Profile, error) {
	return s.repo.GetProfileByID(ctx, id)
}

func (s *service) OnboardStudent(ctx context.Context, profileID string, req *OnboardingRequest) error {
	return s.repo.CompleteStudentOnboarding(ctx, profileID, req)
}

func (s *service) UpdateProfile(ctx context.Context, p *Profile) error {
	return s.repo.UpdateProfile(ctx, p)
}
