package content

import (
	"math"
	"testing"
)

func TestCalculateRank(t *testing.T) {
	tests := []struct {
		name                string
		signals             RankingSignals
		expectedExplanation string
		// We won't test exact float equality because of math.Exp, but we can test relative ordering
	}{
		{
			name: "Roadmap Current - High Priority",
			signals: RankingSignals{
				BaseQuality:  1.0,
				MaxInterest:  0.5,
				RoadmapLevel: 2, // current
				DaysSince:    0,
			},
			expectedExplanation: "Relevant to your current learning roadmap",
		},
		{
			name: "Roadmap Next - Medium Priority",
			signals: RankingSignals{
				BaseQuality:  1.0,
				MaxInterest:  0.5,
				RoadmapLevel: 1, // next
				DaysSince:    0,
			},
			expectedExplanation: "Relevant to your upcoming learning roadmap",
		},
		{
			name: "Strong Interest - No Roadmap",
			signals: RankingSignals{
				BaseQuality:  1.0,
				MaxInterest:  0.8,
				RoadmapLevel: 0,
				DaysSince:    0,
			},
			expectedExplanation: "Strongly matches your interests",
		},
		{
			name: "Weak Interest - No Roadmap",
			signals: RankingSignals{
				BaseQuality:  1.0,
				MaxInterest:  0.3,
				RoadmapLevel: 0,
				DaysSince:    0,
			},
			expectedExplanation: "Matches your interests",
		},
		{
			name: "No Interest - Recent Source",
			signals: RankingSignals{
				BaseQuality:  1.0,
				MaxInterest:  0.0,
				RoadmapLevel: 0,
				DaysSince:    0,
			},
			expectedExplanation: "Recent content from your sources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, explanation := CalculateRank(tt.signals)
			if explanation != tt.expectedExplanation {
				t.Errorf("expected explanation %q, got %q", tt.expectedExplanation, explanation)
			}
		})
	}
}

func TestRankAndSortFeed(t *testing.T) {
	// Item 1: High quality, no interest, very fresh
	item1 := ContentItem{ID: 1, Title: "Item 1"}
	sig1 := RankingSignals{BaseQuality: 1.0, MaxInterest: 0, RoadmapLevel: 0, DaysSince: 0}

	// Item 2: Medium quality, strong interest, kinda fresh
	item2 := ContentItem{ID: 2, Title: "Item 2"}
	sig2 := RankingSignals{BaseQuality: 0.8, MaxInterest: 0.9, RoadmapLevel: 0, DaysSince: 1}

	// Item 3: Low quality, no interest, old
	item3 := ContentItem{ID: 3, Title: "Item 3"}
	sig3 := RankingSignals{BaseQuality: 0.5, MaxInterest: 0, RoadmapLevel: 0, DaysSince: 10}

	// Item 4: High quality, Roadmap Current, old but important
	item4 := ContentItem{ID: 4, Title: "Item 4"}
	sig4 := RankingSignals{BaseQuality: 1.0, MaxInterest: 0.5, RoadmapLevel: 2, DaysSince: 5}

	rawItems := map[int64]ContentItem{
		1: item1,
		2: item2,
		3: item3,
		4: item4,
	}

	signals := map[int64]RankingSignals{
		1: sig1,
		2: sig2,
		3: sig3,
		4: sig4,
	}

	result := RankAndSortFeed(rawItems, signals, 10)

	// We expect a specific deterministic order based on our constants:
	// item4 Base: (1.0*1.0) + (0.5*2.0) + 3.0 = 5.0. Decay: exp(-0.1*5)=0.606. Final: ~3.03
	// item2 Base: (0.8*1.0) + (0.9*2.0) + 0.0 = 2.6. Decay: exp(-0.1*1)=0.904. Final: ~2.35
	// item1 Base: (1.0*1.0) + 0 + 0 = 1.0. Decay: exp(0)=1. Final: 1.0
	// item3 Base: (0.5*1.0) + 0 + 0 = 0.5. Decay: exp(-0.1*10)=0.367. Final: ~0.18

	if len(result) != 4 {
		t.Fatalf("expected 4 items, got %d", len(result))
	}

	expectedOrder := []int64{4, 2, 1, 3}
	for i, expectedID := range expectedOrder {
		if result[i].ID != expectedID {
			t.Errorf("rank position %d: expected item %d, got %d. Score: %f", i, expectedID, result[i].ID, *result[i].RankScore)
		}
	}
}

func TestTimeDecay(t *testing.T) {
	// Two identical items, one is older
	sigFresh := RankingSignals{BaseQuality: 1.0, MaxInterest: 0.5, RoadmapLevel: 0, DaysSince: 0}
	sigOld := RankingSignals{BaseQuality: 1.0, MaxInterest: 0.5, RoadmapLevel: 0, DaysSince: 10}

	scoreFresh, _ := CalculateRank(sigFresh)
	scoreOld, _ := CalculateRank(sigOld)

	if scoreOld >= scoreFresh {
		t.Errorf("expected older item to score less than fresh item. Fresh: %f, Old: %f", scoreFresh, scoreOld)
	}

	// Calculate expected ratio: math.Exp(-0.1 * 10) / math.Exp(0)
	expectedRatio := math.Exp(-1.0)
	actualRatio := scoreOld / scoreFresh

	// Float comparison with small tolerance
	if math.Abs(actualRatio-expectedRatio) > 1e-9 {
		t.Errorf("decay ratio incorrect. Expected %f, got %f", expectedRatio, actualRatio)
	}
}
