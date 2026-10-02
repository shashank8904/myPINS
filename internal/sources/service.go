package sources

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(ctx context.Context, source Source) (Source, error) {
	return s.repository.Create(ctx, source)
}

func (s *Service) GetByID(ctx context.Context, id int64) (Source, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]Source, error) {
	return s.repository.List(ctx)
}
