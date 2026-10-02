package topics

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(ctx context.Context, topic Topic) (Topic, error) {
	return s.repository.Create(ctx, topic)
}

func (s *Service) GetByID(ctx context.Context, id int64) (Topic, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]Topic, error) {
	return s.repository.List(ctx)
}
