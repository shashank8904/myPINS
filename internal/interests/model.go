package interests

import "time"

type Interest struct {
	ID        int64     `json:"id"`
	TopicID   int64     `json:"topic_id"`
	Weight    float64   `json:"weight"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
