package content

import "time"

type ContentItem struct {
	ID           int64
	SourceID     int64
	Title        string
	URL          string
	ContentType  string
	Summary      *string
	WhyItMatters *string
	Body         *string
	Author       *string
	PublishedAt  *time.Time
	IngestedAt   time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Phase 7: Personalization (Not persisted in content_items table)
	RankScore                  *float64
	PersonalizationExplanation *string
}

type ContentItemTopic struct {
	ID             int64
	ContentItemID  int64
	TopicID        int64
	RelevanceScore float64
	CreatedAt      time.Time
}

type UserItemInteraction struct {
	ID            int64
	ContentItemID int64
	Action        string
	InteractedAt  time.Time
}

type RSSItem struct {
	Title       string
	URL         string
	Author      string
	Description string
	PublishedAt time.Time
}
