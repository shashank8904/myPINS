package roadmap

import "time"

type RoadmapItem struct {
	ID        int64     `json:"id"`
	TopicID   int64     `json:"topic_id"`
	Status    string    `json:"status"`
	Priority  *int      `json:"priority"`
	Notes     *string   `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
