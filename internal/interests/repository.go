package interests

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

func (r *Repository) Create(ctx context.Context, interest Interest) (Interest, error) {
	query := `
		INSERT INTO interests (
			topic_id,
			weight
		)
		VALUES ($1, $2)
		RETURNING
			id,
			topic_id,
			weight,
			created_at,
			updated_at
	`

	var created Interest

	err := r.db.QueryRow(
		ctx,
		query,
		interest.TopicID,
		interest.Weight,
	).Scan(
		&created.ID,
		&created.TopicID,
		&created.Weight,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return Interest{}, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (Interest, error) {
	query := `
		SELECT
			id,
			topic_id,
			weight,
			created_at,
			updated_at
		FROM interests
		WHERE id = $1
	`

	var interest Interest

	err := r.db.QueryRow(ctx, query, id).Scan(
		&interest.ID,
		&interest.TopicID,
		&interest.Weight,
		&interest.CreatedAt,
		&interest.UpdatedAt,
	)

	if err != nil {
		return Interest{}, err
	}

	return interest, nil
}

func (r *Repository) List(ctx context.Context) ([]Interest, error) {
	query := `
		SELECT
			id,
			topic_id,
			weight,
			created_at,
			updated_at
		FROM interests
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interests []Interest

	for rows.Next() {
		var interest Interest

		err := rows.Scan(
			&interest.ID,
			&interest.TopicID,
			&interest.Weight,
			&interest.CreatedAt,
			&interest.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		interests = append(interests, interest)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return interests, nil
}
