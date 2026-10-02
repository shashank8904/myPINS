package sources

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

func (r *Repository) Create(ctx context.Context, source Source) (Source, error) {
	query := `
		INSERT INTO sources (
			name,
			url,
			source_type,
			description,
			feed_url
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			name,
			url,
			source_type,
			description,
			feed_url,
			is_active,
			created_at,
			updated_at,
			base_quality_score,
			quality_explanation
	`

	var created Source

	err := r.db.QueryRow(
		ctx,
		query,
		source.Name,
		source.URL,
		source.SourceType,
		source.Description,
		source.FeedURL,
	).Scan(
		&created.ID,
		&created.Name,
		&created.URL,
		&created.SourceType,
		&created.Description,
		&created.FeedURL,
		&created.IsActive,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.BaseQualityScore,
		&created.QualityExplanation,
	)

	if err != nil {
		return Source{}, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (Source, error) {
	query := `
		SELECT
			id,
			name,
			url,
			source_type,
			description,
			feed_url,
			is_active,
			created_at,
			updated_at,
			base_quality_score,
			quality_explanation
		FROM sources
		WHERE id = $1
	`

	var source Source

	err := r.db.QueryRow(ctx, query, id).Scan(
		&source.ID,
		&source.Name,
		&source.URL,
		&source.SourceType,
		&source.Description,
		&source.FeedURL,
		&source.IsActive,
		&source.CreatedAt,
		&source.UpdatedAt,
		&source.BaseQualityScore,
		&source.QualityExplanation,
	)

	if err != nil {
		return Source{}, err
	}

	return source, nil
}

func (r *Repository) List(ctx context.Context) ([]Source, error) {
	query := `
		SELECT
			id,
			name,
			url,
			source_type,
			description,
			feed_url,
			is_active,
			created_at,
			updated_at,
			base_quality_score,
			quality_explanation
		FROM sources
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []Source

	for rows.Next() {
		var source Source

		err := rows.Scan(
			&source.ID,
			&source.Name,
			&source.URL,
			&source.SourceType,
			&source.Description,
			&source.FeedURL,
			&source.IsActive,
			&source.CreatedAt,
			&source.UpdatedAt,
			&source.BaseQualityScore,
			&source.QualityExplanation,
		)
		if err != nil {
			return nil, err
		}

		sources = append(sources, source)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sources, nil
}

// ListActive returns all sources with is_active = true, ordered by id.
// This is the query used by the ingestion worker on every run to discover
// which feeds need to be processed.
func (r *Repository) ListActive(ctx context.Context) ([]Source, error) {
	query := `
		SELECT
			id,
			name,
			url,
			source_type,
			description,
			feed_url,
			is_active,
			created_at,
			updated_at,
			base_quality_score,
			quality_explanation
		FROM sources
		WHERE is_active = true
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []Source

	for rows.Next() {
		var source Source

		err := rows.Scan(
			&source.ID,
			&source.Name,
			&source.URL,
			&source.SourceType,
			&source.Description,
			&source.FeedURL,
			&source.IsActive,
			&source.CreatedAt,
			&source.UpdatedAt,
			&source.BaseQualityScore,
			&source.QualityExplanation,
		)
		if err != nil {
			return nil, err
		}

		sources = append(sources, source)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sources, nil
}
