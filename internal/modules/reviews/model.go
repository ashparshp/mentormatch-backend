package reviews

import "time"

type Review struct {
	ID        string    `json:"id"`
	BookingID string    `json:"booking_id"`
	StudentID string    `json:"student_id"`
	MentorID  string    `json:"mentor_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	
	// Joined fields
	StudentName string `json:"student_name,omitempty"`
}

type CreateReviewRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment"`
}
