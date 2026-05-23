package analysis

import "context"

// Service runs batch analysis over seller data and prepares trigger candidates.
type Service struct{}

// New creates an analysis service instance.
func New() *Service {
	return &Service{}
}

// Run executes a batch analysis job. Business logic will be added later.
func (s *Service) Run(ctx context.Context) error {
	_ = s
	_ = ctx
	return nil
}
