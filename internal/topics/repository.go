package topics

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, topic Topic) (Topic, error) {
	query := `
		INSERT INTO topics (
			name,
			slug,
			description
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			name,
			slug,
			description,
			created_at,
			updated_at
	`

	var created Topic

	err := r.db.QueryRow(
		ctx,
		query,
		topic.Name,
		topic.Slug,
		topic.Description,
	).Scan(
		&created.ID,
		&created.Name,
		&created.Slug,
		&created.Description,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return Topic{}, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (Topic, error) {
	query := `
		SELECT
			id,
			name,
			slug,
			description,
			created_at,
			updated_at
		FROM topics
		WHERE id = $1
	`

	var topic Topic

	err := r.db.QueryRow(ctx, query, id).Scan(
		&topic.ID,
		&topic.Name,
		&topic.Slug,
		&topic.Description,
		&topic.CreatedAt,
		&topic.UpdatedAt,
	)

	if err != nil {
		return Topic{}, err
	}

	return topic, nil
}

func (r *Repository) List(ctx context.Context) ([]Topic, error) {
	query := `
		SELECT
			id,
			name,
			slug,
			description,
			created_at,
			updated_at
		FROM topics
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []Topic

	for rows.Next() {
		var topic Topic

		err := rows.Scan(
			&topic.ID,
			&topic.Name,
			&topic.Slug,
			&topic.Description,
			&topic.CreatedAt,
			&topic.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		topics = append(topics, topic)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return topics, nil
}

// UpsertByName finds a topic by name or creates it if it doesn't exist.
// This is used by the AI enrichment pipeline to automatically create discovered topics.
func (r *Repository) UpsertByName(ctx context.Context, name string) (Topic, error) {
	// Simple slug generation: lower case, replace spaces with hyphens.
	// In a real app we'd use a better slugifier, but this works for V0.
	// However, doing this in SQL using regexp_replace is cleaner.
	query := `
		WITH new_topic AS (
			INSERT INTO topics (name, slug)
			VALUES (
				$1,
				LOWER(REGEXP_REPLACE($1, '\s+', '-', 'g'))
			)
			ON CONFLICT (slug) DO NOTHING
			RETURNING id, name, slug, description, created_at, updated_at
		)
		SELECT id, name, slug, description, created_at, updated_at FROM new_topic
		UNION ALL
		SELECT id, name, slug, description, created_at, updated_at FROM topics
		WHERE name = $1
		LIMIT 1;
	`

	var topic Topic
	err := r.db.QueryRow(ctx, query, name).Scan(
		&topic.ID,
		&topic.Name,
		&topic.Slug,
		&topic.Description,
		&topic.CreatedAt,
		&topic.UpdatedAt,
	)

	return topic, err
}
