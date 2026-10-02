package roadmap

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

func (r *Repository) Create(ctx context.Context, item RoadmapItem) (RoadmapItem, error) {
	query := `
		INSERT INTO roadmap_items (
			topic_id,
			status,
			priority,
			notes
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			topic_id,
			status,
			priority,
			notes,
			created_at,
			updated_at
	`

	var created RoadmapItem

	err := r.db.QueryRow(
		ctx,
		query,
		item.TopicID,
		item.Status,
		item.Priority,
		item.Notes,
	).Scan(
		&created.ID,
		&created.TopicID,
		&created.Status,
		&created.Priority,
		&created.Notes,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return RoadmapItem{}, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (RoadmapItem, error) {
	query := `
		SELECT
			id,
			topic_id,
			status,
			priority,
			notes,
			created_at,
			updated_at
		FROM roadmap_items
		WHERE id = $1
	`

	var item RoadmapItem

	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.TopicID,
		&item.Status,
		&item.Priority,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		return RoadmapItem{}, err
	}

	return item, nil
}

func (r *Repository) List(ctx context.Context) ([]RoadmapItem, error) {
	query := `
		SELECT
			id,
			topic_id,
			status,
			priority,
			notes,
			created_at,
			updated_at
		FROM roadmap_items
		ORDER BY status, priority ASC NULLS LAST, id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []RoadmapItem

	for rows.Next() {
		var item RoadmapItem

		err := rows.Scan(
			&item.ID,
			&item.TopicID,
			&item.Status,
			&item.Priority,
			&item.Notes,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
