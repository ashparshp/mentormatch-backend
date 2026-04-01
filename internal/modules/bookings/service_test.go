package bookings

import (
	"context"
	"testing"
	"time"

	"github.com/ashparshp/mentormatch-backend/internal/modules/mentors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepo implements bookings.Repository
type MockBookingRepo struct {
	mock.Mock
}

func (m *MockBookingRepo) CreateBooking(ctx context.Context, b *Booking) error {
	args := m.Called(ctx, b)
	return args.Error(0)
}

func (m *MockBookingRepo) GetBookingByID(ctx context.Context, id string) (*Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Booking), args.Error(1)
}

func (m *MockBookingRepo) UpdateBookingStatus(ctx context.Context, id string, status BookingStatus, link string) error {
	args := m.Called(ctx, id, status, link)
	return args.Error(0)
}

func (m *MockBookingRepo) ListBookingsByStudent(ctx context.Context, id string) ([]*Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Booking), args.Error(1)
}

func (m *MockBookingRepo) ListBookingsByMentor(ctx context.Context, id string) ([]*Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Booking), args.Error(1)
}

func (m *MockBookingRepo) GetUpcomingSessions(ctx context.Context, minutes int) ([]*Booking, error) {
	args := m.Called(ctx, minutes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Booking), args.Error(1)
}

// MockMentorRepo implements mentors.Repository
type MockMentorRepo struct {
	mock.Mock
}

func (m *MockMentorRepo) GetMentorByID(ctx context.Context, id string) (*mentors.MentorProfile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mentors.MentorProfile), args.Error(1)
}

func (m *MockMentorRepo) GetActiveMentors(ctx context.Context, filter mentors.MentorFilter) ([]*mentors.MentorProfile, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*mentors.MentorProfile), args.Error(1)
}

func (m *MockMentorRepo) UpdateMentorProfile(ctx context.Context, id string, bio *string, hourlyRate float64, services []mentors.OfferedService) error {
	args := m.Called(ctx, id, bio, hourlyRate, services)
	return args.Error(0)
}

func (m *MockMentorRepo) UpdateAvailability(ctx context.Context, id string, slots map[string][]string) error {
	args := m.Called(ctx, id, slots)
	return args.Error(0)
}

func (m *MockMentorRepo) CompleteMentorOnboarding(ctx context.Context, id string, req *mentors.OnboardMentorRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *MockMentorRepo) SaveProtocolTemplates(ctx context.Context, id string, templates []mentors.ProtocolTemplate) error {
	args := m.Called(ctx, id, templates)
	return args.Error(0)
}

func TestCreateBooking(t *testing.T) {
	mockBookingRepo := new(MockBookingRepo)
	mockMentorRepo := new(MockMentorRepo)
	svc := NewService(mockBookingRepo, mockMentorRepo)

	ctx := context.Background()
	studentID := "student-1"
	mentorID := "mentor-1"
	hourlyRate := 100.0

	mentor := &mentors.MentorProfile{
		ProfileID:  mentorID,
		HourlyRate: hourlyRate,
		Status:     mentors.StatusActive,
	}

	startTime := time.Now().Add(time.Hour)
	endTime := startTime.Add(2 * time.Hour) // 2 hours

	t.Run("Successful Booking Creation", func(t *testing.T) {
		mockMentorRepo.On("GetMentorByID", ctx, mentorID).Return(mentor, nil).Once()
		mockBookingRepo.On("CreateBooking", ctx, mock.AnythingOfType("*bookings.Booking")).Return(nil).Once()

		req := &CreateBookingRequest{
			MentorID:  mentorID,
			StartTime: startTime,
			EndTime:   endTime,
		}

		booking, err := svc.CreateBooking(ctx, studentID, req)

		assert.NoError(t, err)
		assert.NotNil(t, booking)
		assert.Equal(t, 200.0, booking.TotalPrice) // 2 hours * 100 rate
		assert.Equal(t, StatusPending, booking.Status)
		
		mockMentorRepo.AssertExpectations(t)
		mockBookingRepo.AssertExpectations(t)
	})

	t.Run("Mentor Not Found", func(t *testing.T) {
		mockMentorRepo.On("GetMentorByID", ctx, "invalid").Return(nil, nil).Once()

		req := &CreateBookingRequest{
			MentorID:  "invalid",
			StartTime: startTime,
			EndTime:   endTime,
		}

		booking, err := svc.CreateBooking(ctx, studentID, req)

		assert.Error(t, err)
		assert.Nil(t, booking)
		assert.Contains(t, err.Error(), "mentor not found")
	})

	t.Run("Invalid Duration", func(t *testing.T) {
		mockMentorRepo.On("GetMentorByID", ctx, mentorID).Return(mentor, nil).Once()

		req := &CreateBookingRequest{
			MentorID:  mentorID,
			StartTime: endTime,
			EndTime:   startTime, // end before start
		}

		booking, err := svc.CreateBooking(ctx, studentID, req)

		assert.Error(t, err)
		assert.Nil(t, booking)
		assert.Contains(t, err.Error(), "invalid duration")
	})
}
