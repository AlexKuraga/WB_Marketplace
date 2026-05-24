package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wb-marketplace/internal/domain"
)

// NotificationRepository loads and updates seller notification logs.
type NotificationRepository interface {
	ListReadyToSend(ctx context.Context, limit int) ([]domain.SellerNotificationLog, error)
	MarkSent(ctx context.Context, notificationID int64) error
}

type postgresNotificationRepository struct {
	pool *pgxpool.Pool
}

// NewNotificationRepository creates a PostgreSQL-backed notification repository.
func NewNotificationRepository(pool *pgxpool.Pool) NotificationRepository {
	return &postgresNotificationRepository{pool: pool}
}

const selectSellerNotificationLogColumns = `
	id, seller_id, recommendation_id, channel_code, delivery_system_id,
	status, payload_json, sent_at, delivered_at, opened_at, clicked_at,
	error_message, created_at, updated_at
`

func (r *postgresNotificationRepository) ListReadyToSend(ctx context.Context, limit int) ([]domain.SellerNotificationLog, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+selectSellerNotificationLogColumns+`
		FROM seller_notification_log
		WHERE status = 'ready_to_send'
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query ready notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]domain.SellerNotificationLog, 0)
	for rows.Next() {
		item, err := scanSellerNotificationLog(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ready notifications: %w", err)
	}

	return notifications, nil
}

func (r *postgresNotificationRepository) MarkSent(ctx context.Context, notificationID int64) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE seller_notification_log
		SET status = 'sent', sent_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, notificationID)
	if err != nil {
		return fmt.Errorf("mark notification sent: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("notification %d not found", notificationID)
	}
	return nil
}

func scanSellerNotificationLog(row pgx.Row) (domain.SellerNotificationLog, error) {
	var item domain.SellerNotificationLog
	err := row.Scan(
		&item.ID,
		&item.SellerID,
		&item.RecommendationID,
		&item.ChannelCode,
		&item.DeliverySystemID,
		&item.Status,
		&item.PayloadJSON,
		&item.SentAt,
		&item.DeliveredAt,
		&item.OpenedAt,
		&item.ClickedAt,
		&item.ErrorMessage,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return domain.SellerNotificationLog{}, fmt.Errorf("scan seller notification log: %w", err)
	}
	return item, nil
}
