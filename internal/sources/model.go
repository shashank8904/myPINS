package sources

import "time"

type Source struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	SourceType  string    `json:"source_type"`
	Description *string   `json:"description"`
	FeedURL     *string   `json:"feed_url"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Phase 3: Source Quality
	BaseQualityScore   float64 `json:"base_quality_score"`
	QualityExplanation *string `json:"quality_explanation,omitempty"`
}
