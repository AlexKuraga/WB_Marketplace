package service

import "wb-marketplace/internal/repository"

// Services groups application use-case services.
type Services struct {
	Recommendation *RecommendationService
}

// New creates the application service layer.
func New(repos repositoryBundle) *Services {
	return &Services{
		Recommendation: NewRecommendationService(repos),
	}
}

type repositoryBundle struct {
	Recommendations repository.RecommendationRepository
}

// Repositories groups data access dependencies for the service layer.
func Repositories(
	recommendations repository.RecommendationRepository,
) repositoryBundle {
	return repositoryBundle{
		Recommendations: recommendations,
	}
}