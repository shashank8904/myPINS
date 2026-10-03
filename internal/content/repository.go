package content

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// errItemNotFound is the sentinel error returned by GetByID when no row exists
// for the requested ID. Callers use errors.Is(err, errItemNotFound) to
// distinguish a missing resource (404) from a real database failure (500).
// It is unexported because it is an implementation detail of this package;
// only the handler layer needs to distinguish these two cases.
var errItemNotFound = fmt.Errorf("content item not found")

// Repository handles persistence for ContentItem records.
// It talks to PostgreSQL directly via a pgx connection, following the same
// pattern used by the sources, topics, interests, and roadmap packages.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// columnList is the canonical ordered list of content_items columns we
// SELECT or RETURNING. Keeping it in one place prevents column-order bugs
// between INSERT RETURNING and SELECT queries.
const columnList = `
	id,
	source_id,
	title,
	url,
	content_type,
	summary,
	why_it_matters,
	body,
	author,
	published_at,
	ingested_at,
	created_at,
	updated_at
`

// scanItem reads a row with the column order defined by columnList into item.
func scanItem(row pgx.Row, item *ContentItem) error {
	return row.Scan(
		&item.ID,
		&item.SourceID,
		&item.Title,
		&item.URL,
		&item.ContentType,
		&item.Summary,
		&item.WhyItMatters,
		&item.Body,
		&item.Author,
		&item.PublishedAt,
		&item.IngestedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
}

// Create inserts a new ContentItem and returns the fully populated row
// (including database-assigned ID and timestamps).
// It returns an error if the URL already exists (unique constraint violation).
// Use CreateOrIgnore when you want silent deduplication instead.
func (r *Repository) Create(ctx context.Context, item ContentItem) (ContentItem, error) {
	query := `
		INSERT INTO content_items (
			source_id,
			title,
			url,
			content_type,
			summary,
			body,
			author,
			published_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING` + columnList

	var created ContentItem
	row := r.db.QueryRow(
		ctx,
		query,
		item.SourceID,
		item.Title,
		item.URL,
		item.ContentType,
		item.Summary,
		item.Body,
		item.Author,
		item.PublishedAt,
	)
	if err := scanItem(row, &created); err != nil {
		return ContentItem{}, err
	}
	return created, nil
}

// CreateOrIgnore attempts to insert a ContentItem.
// If the URL already exists (unique constraint on content_items.url),
// the insert is skipped and the second return value is false.
//
// PostgreSQL handles deduplication atomically via ON CONFLICT (url) DO NOTHING.
// There is no application-level "check then insert" — that would be a race.
//
// Returns (inserted item, true, nil)  when the row was inserted.
// Returns (zero, false, nil)          when the row was skipped (duplicate URL).
// Returns (zero, false, err)          on any other database error.
func (r *Repository) CreateOrIgnore(ctx context.Context, item ContentItem) (ContentItem, bool, error) {
	query := `
		INSERT INTO content_items (
			source_id,
			title,
			url,
			content_type,
			summary,
			body,
			author,
			published_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (url) DO NOTHING
		RETURNING` + columnList

	var created ContentItem
	row := r.db.QueryRow(
		ctx,
		query,
		item.SourceID,
		item.Title,
		item.URL,
		item.ContentType,
		item.Summary,
		item.Body,
		item.Author,
		item.PublishedAt,
	)

	err := scanItem(row, &created)
	if err != nil {
		// pgx.ErrNoRows is returned when ON CONFLICT DO NOTHING skips the row
		// and RETURNING produces no result. This is the expected "duplicate url" path.
		if errors.Is(err, pgx.ErrNoRows) {
			return ContentItem{}, false, nil
		}

		// Phase 3: Layer 2 Deduplication
		// If the URL changed but the source + normalized title is identical,
		// PostgreSQL will throw a unique constraint violation on our new index.
		// We treat this exactly like an ignored URL conflict.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "uq_content_items_source_title" {
			return ContentItem{}, false, nil
		}

		return ContentItem{}, false, err
	}

	return created, true, nil
}

// GetByID retrieves a single ContentItem by its primary key.
// Returns errItemNotFound (check with errors.Is) when no row exists for the
// given ID. All other errors represent database or scanning failures.
func (r *Repository) GetByID(ctx context.Context, id int64) (ContentItem, error) {
	query := `
		SELECT` + columnList + `
		FROM content_items
		WHERE id = $1
	`

	var item ContentItem
	row := r.db.QueryRow(ctx, query, id)
	if err := scanItem(row, &item); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ContentItem{}, errItemNotFound
		}
		return ContentItem{}, err
	}
	return item, nil
}

// List returns all ContentItems ordered by ingestion time descending (newest first).
// V0 has no pagination — that will be added when the feed API is built.
func (r *Repository) List(ctx context.Context) ([]ContentItem, error) {
	query := `
		SELECT` + columnList + `
		FROM content_items
		ORDER BY ingested_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ContentItem
	for rows.Next() {
		var item ContentItem
		if err := scanItem(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// ListPersonalized returns a ranked feed based on user interests, roadmap, and time decay.
// It extracts the raw ranking signals from the database and delegates the actual
// ranking mathematics and sorting to the Go ranking package.
func (r *Repository) ListPersonalized(ctx context.Context, limit int) ([]ContentItem, error) {
	// We fetch up to 1000 unread items to score in memory.
	// This is perfectly safe for a personal V0 and makes the ranking highly testable.
	query := `
		WITH topic_feedback AS (
			SELECT 
				cit.topic_id,
				SUM(
					CASE 
						WHEN uii.action = 'saved' THEN 2.0
						WHEN uii.action = 'read' THEN 1.0
						WHEN uii.action = 'dismissed' THEN -1.0
						ELSE 0.0
					END
				) AS raw_feedback
			FROM user_item_interactions uii
			JOIN content_item_topics cit ON uii.content_item_id = cit.content_item_id
			GROUP BY cit.topic_id
		)
		SELECT 
			ci.id, ci.source_id, ci.title, ci.url, ci.content_type, ci.summary, ci.why_it_matters, ci.body, ci.author, ci.published_at, ci.ingested_at, ci.created_at, ci.updated_at,
			MAX(COALESCE(s.base_quality_score, 0.5)) AS base_quality,
			MAX(COALESCE(i.weight * cit.relevance_score, 0)) AS max_interest,
			MAX(CASE WHEN ro.status = 'current' THEN 2 WHEN ro.status = 'next' THEN 1 ELSE 0 END) AS roadmap_level,
			EXTRACT(EPOCH FROM (NOW() - COALESCE(ci.published_at, ci.ingested_at)))/86400 AS days_since,
			MAX(COALESCE(tf.raw_feedback, 0)) AS max_raw_feedback
		FROM content_items ci
		JOIN sources s ON ci.source_id = s.id
		LEFT JOIN content_item_topics cit ON ci.id = cit.content_item_id
		LEFT JOIN interests i ON cit.topic_id = i.topic_id
		LEFT JOIN roadmap_items ro ON cit.topic_id = ro.topic_id
		LEFT JOIN topic_feedback tf ON cit.topic_id = tf.topic_id
		LEFT JOIN user_item_interactions uii_exclude ON ci.id = uii_exclude.content_item_id AND uii_exclude.action IN ('read', 'dismissed', 'saved')
		WHERE uii_exclude.id IS NULL -- Exclude items already interacted with
		GROUP BY ci.id
		LIMIT 1000
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rawItems := make(map[int64]ContentItem)
	signals := make(map[int64]RankingSignals)

	for rows.Next() {
		var item ContentItem
		var sig RankingSignals

		err := rows.Scan(
			&item.ID,
			&item.SourceID,
			&item.Title,
			&item.URL,
			&item.ContentType,
			&item.Summary,
			&item.WhyItMatters,
			&item.Body,
			&item.Author,
			&item.PublishedAt,
			&item.IngestedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
			&sig.BaseQuality,
			&sig.MaxInterest,
			&sig.RoadmapLevel,
			&sig.DaysSince,
			&sig.MaxRawFeedback,
		)
		if err != nil {
			return nil, err
		}

		rawItems[item.ID] = item
		signals[item.ID] = sig
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return RankAndSortFeed(rawItems, signals, limit), nil
}

// RecordInteraction upserts a user action (read, saved, dismissed) on a ContentItem.
// If the user performs the same action on the same item, interacted_at is refreshed.
func (r *Repository) RecordInteraction(ctx context.Context, itemID int64, action string) error {
	query := `
		INSERT INTO user_item_interactions (content_item_id, action, interacted_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (content_item_id, action) 
		DO UPDATE SET interacted_at = EXCLUDED.interacted_at
	`

	_, err := r.db.Exec(ctx, query, itemID, action)
	return err
}

// GetUnenrichedItems fetches items that haven't been processed by the AI yet.
// For V0, we assume an item is unenriched if why_it_matters is NULL.
func (r *Repository) GetUnenrichedItems(ctx context.Context, limit int) ([]ContentItem, error) {
	query := `
		SELECT` + columnList + `
		FROM content_items
		WHERE why_it_matters IS NULL
		ORDER BY ingested_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ContentItem
	for rows.Next() {
		var item ContentItem
		if err := scanItem(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// UpdateEnrichment updates the item with the AI-generated summary and why_it_matters.
func (r *Repository) UpdateEnrichment(ctx context.Context, itemID int64, summary string, whyItMatters string) error {
	query := `
		UPDATE content_items
		SET summary = $1, why_it_matters = $2, updated_at = NOW()
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, summary, whyItMatters, itemID)
	return err
}

// LinkTopic links a topic to a content item with a relevance score (0.0 - 1.0).
func (r *Repository) LinkTopic(ctx context.Context, itemID int64, topicID int64, relevanceScore float64) error {
	query := `
		INSERT INTO content_item_topics (content_item_id, topic_id, relevance_score)
		VALUES ($1, $2, $3)
		ON CONFLICT (content_item_id, topic_id) 
		DO UPDATE SET relevance_score = EXCLUDED.relevance_score
	`

	_, err := r.db.Exec(ctx, query, itemID, topicID, relevanceScore)
	return err
}
