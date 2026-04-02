package mentors

import (
	"strings"
)

// ExpertiseRates defines a static high-fidelity mapping of expertise to hourly rates.
// In a production environment, this could be stored in the database or managed
// via a dynamic analytics-driven pricing engine.
var ExpertiseRates = map[string]float64{
	"JEE Physics":      2500.0,
	"JEE Chemistry":    2200.0,
	"JEE Maths":        2500.0,
	"NEET Biology":     2000.0,
	"System Design":    3500.0,
	"Cloud Architect":  3000.0,
	"FAANG Prep":       3000.0,
	"GSoC Prep":        1500.0,
	"React":            1800.0,
	"Go":               2200.0,
	"Python":           1500.0,
	"Product Strategy": 2800.0,
	"Case Studies":     2500.0,
}

// DefaultRate is used when no expertise tags match the registry.
const DefaultRate = 1000.0

// CalculateSuggestedRate analyzes a mentor's expertise and returns the highest
// recommended value to ensure market competitiveness.
func CalculateSuggestedRate(expertise []string) float64 {
	if len(expertise) == 0 {
		return DefaultRate
	}

	maxRate := 0.0
	for _, tag := range expertise {
		// Normalization to ensure robust matching
		normalizedTag := strings.TrimSpace(tag)
		if rate, exists := ExpertiseRates[normalizedTag]; exists {
			if rate > maxRate {
				maxRate = rate
			}
		}
	}

	if maxRate == 0.0 {
		return DefaultRate
	}

	return maxRate
}
