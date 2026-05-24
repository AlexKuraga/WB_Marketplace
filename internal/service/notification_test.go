package service

import (
	"context"
	"errors"
	"testing"

	"wb-marketplace/internal/domain"
)

type mockNotificationRepository struct {
	listReadyToSendFn func(ctx context.Context, limit int) ([]domain.SellerNotificationLog, error)
	markSentFn        func(ctx context.Context, notificationID int64) error
}

func (m *mockNotificationRepository) ListReadyToSend(ctx context.Context, limit int) ([]domain.SellerNotificationLog, error) {
	return m.listReadyToSendFn(ctx, limit)
}

func (m *mockNotificationRepository) MarkSent(ctx context.Context, notificationID int64) error {
	return m.markSentFn(ctx, notificationID)
}

func TestProcessReadyNotificationsEmptyList(t *testing.T) {
	svc := NewNotificationService(repositoryBundle{
		Notifications: &mockNotificationRepository{
			listReadyToSendFn: func(ctx context.Context, limit int) ([]domain.SellerNotificationLog, error) {
				return nil, nil
			},
		},
	})

	processed, err := svc.ProcessReadyNotifications(context.Background(), 100)
	if err != nil {
		t.Fatalf("ProcessReadyNotifications() error = %v", err)
	}
	if processed != 0 {
		t.Fatalf("processed = %d, want 0", processed)
	}
}

func TestProcessReadyNotificationsSuccess(t *testing.T) {
	marked := make([]int64, 0)

	svc := NewNotificationService(repositoryBundle{
		Notifications: &mockNotificationRepository{
			listReadyToSendFn: func(ctx context.Context, limit int) ([]domain.SellerNotificationLog, error) {
				return []domain.SellerNotificationLog{
					{ID: 1},
					{ID: 2},
				}, nil
			},
			markSentFn: func(ctx context.Context, notificationID int64) error {
				marked = append(marked, notificationID)
				return nil
			},
		},
	})

	processed, err := svc.ProcessReadyNotifications(context.Background(), 100)
	if err != nil {
		t.Fatalf("ProcessReadyNotifications() error = %v", err)
	}
	if processed != 2 {
		t.Fatalf("processed = %d, want 2", processed)
	}
	if len(marked) != 2 || marked[0] != 1 || marked[1] != 2 {
		t.Fatalf("marked = %v, want [1 2]", marked)
	}
}

func TestProcessReadyNotificationsPartialFailure(t *testing.T) {
	svc := NewNotificationService(repositoryBundle{
		Notifications: &mockNotificationRepository{
			listReadyToSendFn: func(ctx context.Context, limit int) ([]domain.SellerNotificationLog, error) {
				return []domain.SellerNotificationLog{
					{ID: 1},
					{ID: 2},
					{ID: 3},
				}, nil
			},
			markSentFn: func(ctx context.Context, notificationID int64) error {
				if notificationID == 2 {
					return errors.New("update failed")
				}
				return nil
			},
		},
	})

	processed, err := svc.ProcessReadyNotifications(context.Background(), 100)
	if err != nil {
		t.Fatalf("ProcessReadyNotifications() error = %v", err)
	}
	if processed != 2 {
		t.Fatalf("processed = %d, want 2", processed)
	}
}

func TestProcessReadyNotificationsDefaultLimit(t *testing.T) {
	var gotLimit int

	svc := NewNotificationService(repositoryBundle{
		Notifications: &mockNotificationRepository{
			listReadyToSendFn: func(ctx context.Context, limit int) ([]domain.SellerNotificationLog, error) {
				gotLimit = limit
				return []domain.SellerNotificationLog{}, nil
			},
		},
	})

	_, err := svc.ProcessReadyNotifications(context.Background(), 0)
	if err != nil {
		t.Fatalf("ProcessReadyNotifications() error = %v", err)
	}
	if gotLimit != defaultNotificationProcessLimit {
		t.Fatalf("limit = %d, want %d", gotLimit, defaultNotificationProcessLimit)
	}
}
