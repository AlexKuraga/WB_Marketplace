package service

import (
	"context"
	"errors"
	"testing"

	"wb-marketplace/internal/domain"
)

type mockRecommendationRepository struct {
	getActiveBySellerIDFn func(ctx context.Context, sellerID int64) ([]domain.Recommendation, error)
	updateStatusFn        func(ctx context.Context, recommendationID int64, status string) error
	createFeedbackFn      func(ctx context.Context, recommendationID int64, feedbackType string) error
}

func (m *mockRecommendationRepository) GetActiveBySellerID(ctx context.Context, sellerID int64) ([]domain.Recommendation, error) {
	return m.getActiveBySellerIDFn(ctx, sellerID)
}

func (m *mockRecommendationRepository) UpdateStatus(ctx context.Context, recommendationID int64, status string) error {
	return m.updateStatusFn(ctx, recommendationID, status)
}

func (m *mockRecommendationRepository) CreateFeedback(ctx context.Context, recommendationID int64, feedbackType string) error {
	return m.createFeedbackFn(ctx, recommendationID, feedbackType)
}

func TestGetActiveBySellerReturnsEmptySlice(t *testing.T) {
	svc := NewRecommendationService(Repositories(&mockRecommendationRepository{
		getActiveBySellerIDFn: func(ctx context.Context, sellerID int64) ([]domain.Recommendation, error) {
			return nil, nil
		},
	}))

	recs, err := svc.GetActiveBySeller(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetActiveBySeller() error = %v", err)
	}
	if recs == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(recs) != 0 {
		t.Fatalf("len(recs) = %d, want 0", len(recs))
	}
}

func TestGetActiveBySellerValidation(t *testing.T) {
	svc := NewRecommendationService(Repositories(&mockRecommendationRepository{}))

	_, err := svc.GetActiveBySeller(context.Background(), 0)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRecordView(t *testing.T) {
	var gotStatus, gotFeedback string

	svc := NewRecommendationService(Repositories(&mockRecommendationRepository{
		updateStatusFn: func(ctx context.Context, recommendationID int64, status string) error {
			if recommendationID != 42 {
				t.Fatalf("recommendationID = %d, want 42", recommendationID)
			}
			gotStatus = status
			return nil
		},
		createFeedbackFn: func(ctx context.Context, recommendationID int64, feedbackType string) error {
			if recommendationID != 42 {
				t.Fatalf("recommendationID = %d, want 42", recommendationID)
			}
			gotFeedback = feedbackType
			return nil
		},
	}))

	if err := svc.RecordView(context.Background(), 42); err != nil {
		t.Fatalf("RecordView() error = %v", err)
	}
	if gotStatus != "opened" {
		t.Errorf("status = %q, want opened", gotStatus)
	}
	if gotFeedback != "view" {
		t.Errorf("feedback = %q, want view", gotFeedback)
	}
}

func TestRecordAccept(t *testing.T) {
	var gotStatus, gotFeedback string

	svc := NewRecommendationService(Repositories(&mockRecommendationRepository{
		updateStatusFn: func(ctx context.Context, recommendationID int64, status string) error {
			gotStatus = status
			return nil
		},
		createFeedbackFn: func(ctx context.Context, recommendationID int64, feedbackType string) error {
			gotFeedback = feedbackType
			return nil
		},
	}))

	if err := svc.RecordAccept(context.Background(), 1); err != nil {
		t.Fatalf("RecordAccept() error = %v", err)
	}
	if gotStatus != "accepted" {
		t.Errorf("status = %q, want accepted", gotStatus)
	}
	if gotFeedback != "accept" {
		t.Errorf("feedback = %q, want accept", gotFeedback)
	}
}

func TestRecordReject(t *testing.T) {
	var gotStatus, gotFeedback string

	svc := NewRecommendationService(Repositories(&mockRecommendationRepository{
		updateStatusFn: func(ctx context.Context, recommendationID int64, status string) error {
			gotStatus = status
			return nil
		},
		createFeedbackFn: func(ctx context.Context, recommendationID int64, feedbackType string) error {
			gotFeedback = feedbackType
			return nil
		},
	}))

	if err := svc.RecordReject(context.Background(), 1); err != nil {
		t.Fatalf("RecordReject() error = %v", err)
	}
	if gotStatus != "rejected" {
		t.Errorf("status = %q, want rejected", gotStatus)
	}
	if gotFeedback != "reject" {
		t.Errorf("feedback = %q, want reject", gotFeedback)
	}
}

func TestRecordViewNotFound(t *testing.T) {
	svc := NewRecommendationService(Repositories(&mockRecommendationRepository{
		updateStatusFn: func(ctx context.Context, recommendationID int64, status string) error {
			return errors.New("recommendation 99 not found")
		},
	}))

	err := svc.RecordView(context.Background(), 99)
	if !errors.Is(err, ErrRecommendationNotFound) {
		t.Fatalf("error = %v, want ErrRecommendationNotFound", err)
	}
}
