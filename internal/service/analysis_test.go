package service

import (
	"context"
	"testing"

	"wb-marketplace/internal/domain"
)

type mockAnalysisRepository struct {
	listSellersFn      func(ctx context.Context) ([]domain.Seller, error)
	getLatestSnapshotFn func(ctx context.Context, sellerID int64) (*domain.SellerMetricsSnapshot, error)
	createJobFn        func(ctx context.Context, job domain.AnalysisJob) (int64, error)
	updateJobFn        func(ctx context.Context, job domain.AnalysisJob) error
}

func (m *mockAnalysisRepository) ListSellersForAnalysis(ctx context.Context) ([]domain.Seller, error) {
	return m.listSellersFn(ctx)
}

func (m *mockAnalysisRepository) GetLatestSnapshot(ctx context.Context, sellerID int64) (*domain.SellerMetricsSnapshot, error) {
	return m.getLatestSnapshotFn(ctx, sellerID)
}

func (m *mockAnalysisRepository) CreateTrigger(ctx context.Context, trigger domain.SellerTriggerLog) error {
	return nil
}

func (m *mockAnalysisRepository) CreateRecommendation(ctx context.Context, recommendation domain.Recommendation) error {
	return nil
}

func (m *mockAnalysisRepository) CreateNotificationLog(ctx context.Context, notification domain.SellerNotificationLog) error {
	return nil
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
					return &domain.SellerMetricsSnapshot{ID: 100, SellerID: 10}, nil
				}
				return nil, nil
			},
		},
	})

	if err := svc.RunManualAnalysis(context.Background()); err != nil {
		t.Fatalf("RunManualAnalysis() error = %v", err)
	}
	if createdJob.Status != "running" {
		t.Errorf("created status = %q, want running", createdJob.Status)
	}
	if createdJob.JobType != manualAnalysisJobType {
		t.Errorf("created job type = %q, want %q", createdJob.JobType, manualAnalysisJobType)
	}
	if updatedJob.Status != "success" {
		t.Errorf("updated status = %q, want success", updatedJob.Status)
	}
	if updatedJob.SellersProcessed != 1 {
		t.Errorf("sellers processed = %d, want 1", updatedJob.SellersProcessed)
	}
	if updatedJob.FinishedAt == nil {
		t.Fatal("expected finished_at to be set")
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
