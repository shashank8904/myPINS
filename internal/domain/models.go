// Package domain defines the core data model structs for myPersonalTechRadar.
// These structs are a faithful representation of the V0 database schema and
// are intentionally free of any persistence, HTTP, or business logic.
//
// Naming convention mirrors the SQL table names (snake_case → PascalCase).
// Nullable columns use pointer types so the zero value is distinguishable from
// an absent value (e.g. *string for optional text, *time.Time for optional timestamps).
package domain

import (
	"time"
)

// ─── Source ──────────────────────────────────────────────────────────────────

// SourceType enumerates the kinds of sources the platform can ingest from.
type SourceType string

const (
	SourceTypeRSSFeed    SourceType = "rss_feed"
	SourceTypeGitHub     SourceType = "github"
	SourceTypeYouTube    SourceType = "youtube"
	SourceTypeHackerNews SourceType = "hackernews"
	SourceTypeConference SourceType = "conference"
	SourceTypeOther      SourceType = "other"
)

// Source represents where content originates (a blog, GitHub, YouTube, etc.).
type Source struct {
	ID          int64      `db:"id"`
	Name        string     `db:"name"`
	URL         string     `db:"url"`
	SourceType  SourceType `db:"source_type"`
	Description *string    `db:"description"`
	IsActive    bool       `db:"is_active"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

// ─── Topic ───────────────────────────────────────────────────────────────────

// Topic represents a technology or subject area (Go, Kubernetes, Rust, etc.).
// It is the central concept linking content, interests, and the roadmap.
type Topic struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Slug        string    `db:"slug"` // URL-safe, e.g. "distributed-systems"
	Description *string   `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// ─── ContentItem ─────────────────────────────────────────────────────────────

// ContentType enumerates the structural kinds of content the platform handles.
type ContentType string

const (
	ContentTypeArticle              ContentType = "article"
	ContentTypeGitHubRelease        ContentType = "github_release"
	ContentTypeYouTubeVideo         ContentType = "youtube_video"
	ContentTypeConferenceTalk       ContentType = "conference_talk"
	ContentTypeResearchPaper        ContentType = "research_paper"
	ContentTypeHackerNewsDiscussion ContentType = "hackernews_discussion"
	ContentTypeOther                ContentType = "other"
)

// ContentItem represents a single piece of technology content from a Source.
// It is generic enough to represent articles, releases, videos, talks, papers,
// and discussions.
type ContentItem struct {
	ID           int64       `db:"id"`
	SourceID     int64       `db:"source_id"`
	Title        string      `db:"title"`
	URL          string      `db:"url"`
	ContentType  ContentType `db:"content_type"`
	Summary      *string     `db:"summary"`
	WhyItMatters *string     `db:"why_it_matters"`
	Body         *string     `db:"body"`
	Author       *string     `db:"author"`
	PublishedAt  *time.Time  `db:"published_at"`
	IngestedAt   time.Time   `db:"ingested_at"`
	CreatedAt    time.Time   `db:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at"`
}

// ─── ContentItemTopic ─────────────────────────────────────────────────────────

// ContentItemTopic is the junction entity between ContentItem and Topic.
// RelevanceScore indicates how strongly the item relates to the topic.
// Stored as float64; DB column is NUMERIC(4,3), range 0.000–1.000.
type ContentItemTopic struct {
	ID             int64     `db:"id"`
	ContentItemID  int64     `db:"content_item_id"`
	TopicID        int64     `db:"topic_id"`
	RelevanceScore float64   `db:"relevance_score"` // 0.000 – 1.000
	CreatedAt      time.Time `db:"created_at"`
}

// ─── Interest ────────────────────────────────────────────────────────────────

// Interest captures the user's explicit interest level in a Topic.
// Weight ranges from 0.0 (no interest) to 1.0 (deeply interested).
//
// Note: Interest is conceptually distinct from Topic.
//   - Topic    → describes what a piece of content is about.
//   - Interest → describes how much the user cares about that topic.
type Interest struct {
	ID        int64     `db:"id"`
	TopicID   int64     `db:"topic_id"`
	Weight    float64   `db:"weight"` // 0.000 – 1.000
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// ─── RoadmapItem ─────────────────────────────────────────────────────────────

// RoadmapStatus enumerates where a topic sits in the user's learning journey.
type RoadmapStatus string

const (
	RoadmapStatusCurrent   RoadmapStatus = "current"
	RoadmapStatusNext      RoadmapStatus = "next"
	RoadmapStatusExploring RoadmapStatus = "exploring"
	RoadmapStatusCompleted RoadmapStatus = "completed"
	RoadmapStatusPaused    RoadmapStatus = "paused"
)

// RoadmapItem represents something the user is currently learning or plans to learn.
type RoadmapItem struct {
	ID        int64         `db:"id"`
	TopicID   int64         `db:"topic_id"`
	Status    RoadmapStatus `db:"status"`
	Notes     *string       `db:"notes"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
}

// ─── UserItemInteraction ──────────────────────────────────────────────────────

// InteractionAction enumerates the ways the user can interact with a ContentItem.
type InteractionAction string

const (
	InteractionActionRead      InteractionAction = "read"
	InteractionActionSaved     InteractionAction = "saved"
	InteractionActionDismissed InteractionAction = "dismissed"
)

// UserItemInteraction records a user action on a ContentItem.
// The UNIQUE(content_item_id, action) constraint means one record per
// (item, action) pair; use an upsert to refresh InteractedAt.
//
// V1 note: add UserID int64 here when multi-user support is introduced.
type UserItemInteraction struct {
	ID            int64             `db:"id"`
	ContentItemID int64             `db:"content_item_id"`
	Action        InteractionAction `db:"action"`
	InteractedAt  time.Time         `db:"interacted_at"`
}
