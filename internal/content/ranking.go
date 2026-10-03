package content

import (
	"math"
	"sort"
)

// Ranking weights and constants for the V0 feed algorithm.
const (
	WeightSourceQuality = 1.0
	WeightInterest      = 2.0
	BonusRoadmapCurrent = 3.0
	BonusRoadmapNext    = 2.0
	TimeDecayFactor     = -0.1
)

// RankingSignals holds the raw inputs needed to rank a single content item.
type RankingSignals struct {
	BaseQuality    float64
	MaxInterest    float64
	RoadmapLevel   int     // 2 = 'current', 1 = 'next', 0 = none
	DaysSince      float64 // Days since published (or ingested, if published is null)
	MaxRawFeedback float64 // Raw unscaled sum of historical topic interactions
}

// ScoredItem is a ContentItem enriched with its calculated ranking score.
type ScoredItem struct {
	Item                       ContentItem
	RankScore                  float64
	PersonalizationExplanation string
}

// CalculateRank computes the deterministic rank score for a set of signals.
// It also provides a human-readable explanation for why the item was ranked highly.
func CalculateRank(signals RankingSignals) (score float64, explanation string) {
	behavioralBonus := CalculateBehavioralFeedback(signals.MaxRawFeedback)

	roadmapBonus := 0.0
	if signals.RoadmapLevel >= 2 {
		roadmapBonus = BonusRoadmapCurrent
		explanation = "Relevant to your current learning roadmap"
	} else if signals.RoadmapLevel == 1 {
		roadmapBonus = BonusRoadmapNext
		explanation = "Relevant to your upcoming learning roadmap"
	} else if behavioralBonus > 0.5 && behavioralBonus > signals.MaxInterest {
		explanation = "Based on your recent reading habits"
	} else if signals.MaxInterest > 0.7 {
		explanation = "Strongly matches your interests"
	} else if signals.MaxInterest > 0 {
		explanation = "Matches your interests"
	} else {
		explanation = "Recent content from your sources"
	}

	baseScore := (signals.BaseQuality * WeightSourceQuality) + (signals.MaxInterest * WeightInterest) + roadmapBonus + behavioralBonus

	timeDecayMultiplier := math.Exp(TimeDecayFactor * signals.DaysSince)

	score = baseScore * timeDecayMultiplier
	return score, explanation
}

// RankAndSortFeed takes a map of ContentItems and their signals, calculates the score
// for each, and sorts them descending by score. It returns the top N items up to limit.
func RankAndSortFeed(rawItems map[int64]ContentItem, signals map[int64]RankingSignals, limit int) []ContentItem {
	var scoredItems []ScoredItem

	for id, item := range rawItems {
		sig := signals[id]
		score, explanation := CalculateRank(sig)

		// Map the rank back onto the item pointers if they are needed by the frontend API.
		// Since we are copying the item, we create fresh pointers.
		rankedItem := item

		finalScore := score // create a local copy so the pointer doesn't escape the loop sharing the same address
		rankedItem.RankScore = &finalScore

		finalExplanation := explanation
		rankedItem.PersonalizationExplanation = &finalExplanation

		scoredItems = append(scoredItems, ScoredItem{
			Item:                       rankedItem,
			RankScore:                  score,
			PersonalizationExplanation: explanation,
		})
	}

	// Sort descending by RankScore
	sort.SliceStable(scoredItems, func(i, j int) bool {
		return scoredItems[i].RankScore > scoredItems[j].RankScore
	})

	var result []ContentItem
	for i, si := range scoredItems {
		if i >= limit {
			break
		}
		result = append(result, si.Item)
	}

	return result
}
