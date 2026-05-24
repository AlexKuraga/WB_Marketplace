package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"wb-marketplace/internal/domain"
)

// ErrRecommendationNotFound is returned when a recommendation does not exist.
var ErrRecommendationNotFound = errors.New("recommendation not found")

// RecommendationService implements recommendation-related use cases.
type RecommendationService struct {
	repos repositoryBundle
}

// NewRecommendationService creates a recommendation use-case service.
func NewRecommendationService(repos repositoryBundle) *RecommendationService {
	return &RecommendationService{repos: repos}
}

// GetActiveBySeller returns active recommendations for a seller.
func (s *RecommendationService) GetActiveBySeller(ctx context.Context, sellerID int64) ([]domain.Recommendation, error) {
	if sellerID <= 0 {
		return nil, fmt.Errorf("seller id must be positive")
	}
	if s.repos.Recommendations == nil {
		return []domain.Recommendation{}, nil
	}

	recs, err := s.repos.Recommendations.GetActiveBySellerID(ctx, sellerID)
	if err != nil {
		return nil, err
	}
	if recs == nil {
		return []domain.Recommendation{}, nil
	}
	return recs, nil
}

// RecordView marks a recommendation as opened and records view feedback.
func (s *RecommendationService) RecordView(ctx context.Context, recommendationID int64) error {
	return s.recordAction(ctx, recommendationID, "opened", "view")
}

// RecordAccept marks a recommendation as accepted and records accept feedback.
func (s *RecommendationService) RecordAccept(ctx context.Context, recommendationID int64) error {
	return s.recordAction(ctx, recommendationID, "accepted", "accept")
}

// RecordReject marks a recommendation as rejected and records reject feedback.
func (s *RecommendationService) RecordReject(ctx context.Context, recommendationID int64) error {
	return s.recordAction(ctx, recommendationID, "rejected", "reject")
}

func (s *RecommendationService) recordAction(ctx context.Context, recommendationID int64, status, feedbackType string) error {
	if recommendationID <= 0 {
		return fmt.Errorf("recommendation id must be positive")
	}
	if s.repos.Recommendations == nil {
		return errors.New("recommendation repository not configured")
	}

	if err := s.repos.Recommendations.UpdateStatus(ctx, recommendationID, status); err != nil {
		if isRecommendationNotFound(err) {
			return ErrRecommendationNotFound
		}
		return err
	}

	if err := s.repos.Recommendations.CreateFeedback(ctx, recommendationID, feedbackType); err != nil {
		if isRecommendationNotFound(err) {
			return ErrRecommendationNotFound
		}
		return err
	}

	return nil
}

func isRecommendationNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}
