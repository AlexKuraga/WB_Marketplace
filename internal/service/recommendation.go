package service

import (
	"context"

	"wb-marketplace/internal/domain"
)

// RecommendationService implements recommendation-related use cases.
type RecommendationService struct {
	repos repositoryBundle
}

// NewRecommendationService creates a recommendation use-case service.
func NewRecommendationService(repos repositoryBundle) *RecommendationService {
	return &RecommendationService{repos: repos}
}

// ListActiveBySeller returns active recommendations for a seller.
func (s *RecommendationService) ListActiveBySeller(ctx context.Context, sellerID int64) ([]domain.Recommendation, error) {
	if s.repos.Recommendations == nil {
		return []domain.Recommendation{}, nil
	}
	return s.repos.Recommendations.GetActiveBySellerID(ctx, sellerID)
}