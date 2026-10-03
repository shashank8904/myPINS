package content

import (
	"math"
	"testing"
)

func TestCalculateBehavioralFeedback(t *testing.T) {
	tests := []struct {
		name     string
		rawScore float64
		expected float64
	}{
		{
			name:     "Zero interactions",
			rawScore: 0.0,
			expected: 0.0,
		},
		{
			name:     "One read",
			rawScore: ActionWeightRead,
			expected: 0.1,
		},
		{
			name:     "One save",
			rawScore: ActionWeightSaved,
			expected: 0.2,
		},
		{
			name:     "One dismiss",
			rawScore: ActionWeightDismiss,
			expected: -0.1,
		},
		{
			name:     "Repeated positive accumulates (5 saves, 2 reads)",
			rawScore: (5 * ActionWeightSaved) + (2 * ActionWeightRead), // 12.0
			expected: 1.2,
		},
		{
			name:     "Repeated negative accumulates (8 dismisses)",
			rawScore: 8 * ActionWeightDismiss, // -8.0
			expected: -0.8,
		},
		{
			name:     "Positive bound ceiling hit (30 saves)",
			rawScore: 30 * ActionWeightSaved, // 60.0
			expected: 2.0,                    // Max bounded to 2.0
		},
		{
			name:     "Negative bound floor hit (50 dismisses)",
			rawScore: 50 * ActionWeightDismiss, // -50.0
			expected: -2.0,                     // Min bounded to -2.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateBehavioralFeedback(tt.rawScore)
			if math.Abs(got-tt.expected) > 1e-9 {
				t.Errorf("expected %f, got %f", tt.expected, got)
			}
		})
	}
}
