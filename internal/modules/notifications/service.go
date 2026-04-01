package notifications

import (
	"context"
	"log"
	"time"

	"github.com/ashparshp/mentormatch-backend/internal/modules/bookings"
)

type Service struct {
	bookingRepo bookings.Repository
}

func NewService(bookingRepo bookings.Repository) *Service {
	return &Service{bookingRepo: bookingRepo}
}

// StartReminderTicker starts a background worker that polls for upcoming sessions
// and logs reminder intents. In a production environment, this would integrate
// with AWS SES, Twilio, or Firebase Cloud Messaging.
func (s *Service) StartReminderTicker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("[NOTIFICATIONS] Reminder engine synchronized. Polling every %v", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("[NOTIFICATIONS] Reminder engine shutting down...")
			return
		case <-ticker.C:
			s.processReminders(ctx)
		}
	}
}

func (s *Service) processReminders(ctx context.Context) {
	// Look for sessions starting in the next 30 minutes
	upcoming, err := s.bookingRepo.GetUpcomingSessions(ctx, 30)
	if err != nil {
		log.Printf("[NOTIFICATIONS] Error fetching upcoming sessions: %v", err)
		return
	}

	if len(upcoming) == 0 {
		return
	}

	for _, b := range upcoming {
		// Logic: Log the reminder. This serves as a high-fidelity placeholder
		// for real notification dispatch.
		log.Printf("[ATTENDANCE_REMINDER] Session #%s starting soon. Student: %s | Mentor: %s | Time: %s", 
			b.ID, b.StudentID, b.MentorID, b.StartTime.Format("15:04 MST"))
	}
}
