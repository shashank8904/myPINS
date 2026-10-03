package content

import (
	"math"
)

// Constants for behavioral feedback scoring.
// These weights determine how much each action contributes to the raw score.
const (
	ActionWeightSaved   = 2.0  // Strong positive
	ActionWeightRead    = 1.0  // Weak positive
	ActionWeightDismiss = -1.0 // Negative
)

// CalculateBehavioralFeedback applies a squashing function to a raw feedback score.
// This ensures that while repeated behavior accumulates, an isolated action
// doesn't dominate, and a heavily interacted topic doesn't grow to infinity.
//
// Formula: math.Max(-2.0, math.Min(2.0, rawScore * 0.1))
func CalculateBehavioralFeedback(rawScore float64) float64 {
	// Multiply by 0.1 so it takes e.g. 10 reads to get a full 1.0 bonus.
	scaled := rawScore * 0.1

	// Bound between -2.0 and +2.0
	return math.Max(-2.0, math.Min(2.0, scaled))
}
