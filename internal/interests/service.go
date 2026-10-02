package interests

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(ctx context.Context, interest Interest) (Interest, error) {
	return s.repository.Create(ctx, interest)
}

func (s *Service) GetByID(ctx context.Context, id int64) (Interest, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]Interest, error) {
	return s.repository.List(ctx)
}
