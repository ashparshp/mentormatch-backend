package recommendations

import (
	"github.com/ashparshp/mentormatch-backend/internal/modules/mentors"
)

type Recommendation struct {
	Mentor         *mentors.MentorProfile `json:"mentor"`
	MatchScore     float64                `json:"match_score"`      // 0.0 to 1.0 (Percentage)
	MatchReason    string                 `json:"match_reason"`     // e.g. "Matched for: React, Go"
	CommonInterests []string              `json:"common_interests"` // Shared tags
}

type RecommendationResponse struct {
	Recommendations []*Recommendation `json:"recommendations"`
	TotalCount      int               `json:"total_count"`
}
