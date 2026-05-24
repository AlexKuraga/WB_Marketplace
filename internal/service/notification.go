package service

import (
	"context"
	"errors"
	"log"

	"wb-marketplace/internal/domain"
)

const defaultNotificationProcessLimit = 100

// NotificationService processes pending seller notifications.
type NotificationService struct {
	repos repositoryBundle
}

// NewNotificationService creates a notification service instance.
func NewNotificationService(repos repositoryBundle) *NotificationService {
	return &NotificationService{repos: repos}
}

// ProcessReadyNotifications marks ready_to_send notifications as sent.
func (s *NotificationService) ProcessReadyNotifications(ctx context.Context, limit int) (int, error) {
	if s.repos.Notifications == nil {
		return 0, errors.New("notification repository not configured")
	}
	if limit <= 0 {
		limit = defaultNotificationProcessLimit
	}

	notifications, err := s.repos.Notifications.ListReadyToSend(ctx, limit)
	if err != nil {
		return 0, err
	}
	if notifications == nil {
		notifications = make([]domain.SellerNotificationLog, 0)
	}

	processed := 0
	for _, notification := range notifications {
		if err := s.repos.Notifications.MarkSent(ctx, notification.ID); err != nil {
			log.Printf("failed to mark notification %d as sent: %v", notification.ID, err)
			continue
		}
		processed++
	}

	return processed, nil
}
