package roadmap

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(ctx context.Context, item RoadmapItem) (RoadmapItem, error) {
	return s.repository.Create(ctx, item)
}

func (s *Service) GetByID(ctx context.Context, id int64) (RoadmapItem, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]RoadmapItem, error) {
	return s.repository.List(ctx)
}
