package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"wb-marketplace/internal/domain"
)

type mockAnalysisRepository struct {
	listSellersFn       func(ctx context.Context) ([]domain.Seller, error)
	getLatestSnapshotFn func(ctx context.Context, sellerID int64) (*domain.SellerMetricsSnapshot, error)
	createTriggerFn     func(ctx context.Context, trigger domain.SellerTriggerLog) (int64, error)
	createRecommendationFn func(ctx context.Context, recommendation domain.Recommendation) (int64, error)
	createNotificationFn func(ctx context.Context, notification domain.SellerNotificationLog) error
	createJobFn         func(ctx context.Context, job domain.AnalysisJob) (int64, error)
	updateJobFn         func(ctx context.Context, job domain.AnalysisJob) error
}

func (m *mockAnalysisRepository) ListSellersForAnalysis(ctx context.Context) ([]domain.Seller, error) {
	return m.listSellersFn(ctx)
}

func (m *mockAnalysisRepository) GetLatestSnapshot(ctx context.Context, sellerID int64) (*domain.SellerMetricsSnapshot, error) {
	return m.getLatestSnapshotFn(ctx, sellerID)
}

func (m *mockAnalysisRepository) CreateTrigger(ctx context.Context, trigger domain.SellerTriggerLog) (int64, error) {
	return m.createTriggerFn(ctx, trigger)
}

func (m *mockAnalysisRepository) CreateRecommendation(ctx context.Context, recommendation domain.Recommendation) (int64, error) {
	return m.createRecommendationFn(ctx, recommendation)
}

func (m *mockAnalysisRepository) CreateNotificationLog(ctx context.Context, notification domain.SellerNotificationLog) error {
	return m.createNotificationFn(ctx, notification)
}

func (m *mockAnalysisRepository) CreateAnalysisJob(ctx context.Context, job domain.AnalysisJob) (int64, error) {
	return m.createJobFn(ctx, job)
}

func (m *mockAnalysisRepository) UpdateAnalysisJob(ctx context.Context, job domain.AnalysisJob) error {
	return m.updateJobFn(ctx, job)
}

func TestRunManualAnalysisSuccess(t *testing.T) {
	var createdJob domain.AnalysisJob
	var updatedJob domain.AnalysisJob

	svc := NewAnalysisService(repositoryBundle{
		Analysis: &mockAnalysisRepository{
			createJobFn: func(ctx context.Context, job domain.AnalysisJob) (int64, error) {
				createdJob = job
				return 1, nil
			},
			updateJobFn: func(ctx context.Context, job domain.AnalysisJob) error {
				updatedJob = job
				return nil
			},
			listSellersFn: func(ctx context.Context) ([]domain.Seller, error) {
				return []domain.Seller{{ID: 10}, {ID: 20}}, nil
			},
			getLatestSnapshotFn: func(ctx context.Context, sellerID int64) (*domain.SellerMetricsSnapshot, error) {
				if sellerID == 10 {
					return &domain.SellerMetricsSnapshot{
						ID:                  100,
						SellerID:            10,
						SnapshotDate:        time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
						ActiveProductsCount: 1,
					}, nil
				}
				return nil, nil
			},
			createTriggerFn: func(ctx context.Context, trigger domain.SellerTriggerLog) (int64, error) {
				return 0, nil
			},
			createRecommendationFn: func(ctx context.Context, recommendation domain.Recommendation) (int64, error) {
				return 0, nil
			},
			createNotificationFn: func(ctx context.Context, notification domain.SellerNotificationLog) error {
				return nil
			},
		},
	})

	if err := svc.RunManualAnalysis(context.Background()); err != nil {
		t.Fatalf("RunManualAnalysis() error = %v", err)
	}
	if createdJob.Status != "running" {
		t.Errorf("created status = %q, want running", createdJob.Status)
	}
	if updatedJob.Status != "success" {
		t.Errorf("updated status = %q, want success", updatedJob.Status)
	}
	if updatedJob.SellersProcessed != 1 {
		t.Errorf("sellers processed = %d, want 1", updatedJob.SellersProcessed)
	}
}

func TestRunManualAnalysisCreatesRecommendations(t *testing.T) {
	var recommendations []domain.Recommendation
	var notifications []domain.SellerNotificationLog

	svc := NewAnalysisService(repositoryBundle{
		Analysis: &mockAnalysisRepository{
			createJobFn: func(ctx context.Context, job domain.AnalysisJob) (int64, error) {
				return 1, nil
			},
			updateJobFn: func(ctx context.Context, job domain.AnalysisJob) error {
				return nil
			},
			listSellersFn: func(ctx context.Context) ([]domain.Seller, error) {
				return []domain.Seller{{ID: 10}}, nil
			},
			getLatestSnapshotFn: func(ctx context.Context, sellerID int64) (*domain.SellerMetricsSnapshot, error) {
				return &domain.SellerMetricsSnapshot{
					ID:                  100,
					SellerID:            10,
					SnapshotDate:        time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
					ActiveProductsCount: 0,
				}, nil
			},
			createTriggerFn: func(ctx context.Context, trigger domain.SellerTriggerLog) (int64, error) {
				if trigger.TriggerCode != "NO_PRODUCTS" {
					t.Errorf("TriggerCode = %q, want NO_PRODUCTS", trigger.TriggerCode)
				}
				return 501, nil
			},
			createRecommendationFn: func(ctx context.Context, recommendation domain.Recommendation) (int64, error) {
				recommendations = append(recommendations, recommendation)
				return 601, nil
			},
			createNotificationFn: func(ctx context.Context, notification domain.SellerNotificationLog) error {
				notifications = append(notifications, notification)
				return nil
			},
		},
	})

	if err := svc.RunManualAnalysis(context.Background()); err != nil {
		t.Fatalf("RunManualAnalysis() error = %v", err)
	}
	if len(recommendations) != 1 {
		t.Fatalf("len(recommendations) = %d, want 1", len(recommendations))
	}
	if recommendations[0].TriggerID != 501 {
		t.Errorf("TriggerID = %d, want 501", recommendations[0].TriggerID)
	}
	if recommendations[0].Title != "Добавьте первые товары" {
		t.Errorf("Title = %q", recommendations[0].Title)
	}
	if recommendations[0].Status != "created" {
		t.Errorf("Status = %q, want created", recommendations[0].Status)
	}
	if len(notifications) != 1 {
		t.Fatalf("len(notifications) = %d, want 1", len(notifications))
	}
	if notifications[0].RecommendationID != 601 {
		t.Errorf("RecommendationID = %d, want 601", notifications[0].RecommendationID)
	}
	if notifications[0].Status != "ready_to_send" {
		t.Errorf("Status = %q, want ready_to_send", notifications[0].Status)
	}
	if notifications[0].ChannelCode != "in_app" {
		t.Errorf("ChannelCode = %q, want in_app", notifications[0].ChannelCode)
	}
}

func TestRunManualAnalysisSkipsDuplicateTrigger(t *testing.T) {
	createRecommendationCalled := false

	svc := NewAnalysisService(repositoryBundle{
		Analysis: &mockAnalysisRepository{
			createJobFn: func(ctx context.Context, job domain.AnalysisJob) (int64, error) {
				return 1, nil
			},
			updateJobFn: func(ctx context.Context, job domain.AnalysisJob) error {
				return nil
			},
			listSellersFn: func(ctx context.Context) ([]domain.Seller, error) {
				return []domain.Seller{{ID: 10}}, nil
			},
			getLatestSnapshotFn: func(ctx context.Context, sellerID int64) (*domain.SellerMetricsSnapshot, error) {
				return &domain.SellerMetricsSnapshot{
					ID:                  100,
					SellerID:            10,
					SnapshotDate:        time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
					ActiveProductsCount: 0,
				}, nil
			},
			createTriggerFn: func(ctx context.Context, trigger domain.SellerTriggerLog) (int64, error) {
				return 0, &pgconn.PgError{Code: "23505"}
			},
			createRecommendationFn: func(ctx context.Context, recommendation domain.Recommendation) (int64, error) {
				createRecommendationCalled = true
				return 0, nil
			},
			createNotificationFn: func(ctx context.Context, notification domain.SellerNotificationLog) error {
				return nil
			},
		},
	})

	if err := svc.RunManualAnalysis(context.Background()); err != nil {
		t.Fatalf("RunManualAnalysis() error = %v", err)
	}
	if createRecommendationCalled {
		t.Fatal("expected duplicate trigger to skip recommendation creation")
	}
}

func TestRunManualAnalysisListSellersError(t *testing.T) {
	svc := NewAnalysisService(repositoryBundle{
		Analysis: &mockAnalysisRepository{
			createJobFn: func(ctx context.Context, job domain.AnalysisJob) (int64, error) {
				return 1, nil
			},
			updateJobFn: func(ctx context.Context, job domain.AnalysisJob) error {
				if job.Status != "failed" {
					t.Errorf("status = %q, want failed", job.Status)
				}
				return nil
			},
			listSellersFn: func(ctx context.Context) ([]domain.Seller, error) {
				return nil, context.Canceled
			},
		},
	})

	err := svc.RunManualAnalysis(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIsDuplicateTriggerError(t *testing.T) {
	if !isDuplicateTriggerError(&pgconn.PgError{Code: "23505"}) {
		t.Fatal("expected duplicate trigger error")
	}
	if isDuplicateTriggerError(errors.New("other error")) {
		t.Fatal("did not expect duplicate trigger error")
	}
}
