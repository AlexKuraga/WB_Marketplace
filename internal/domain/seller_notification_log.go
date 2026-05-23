package domain

import (
	"encoding/json"
	"time"
)

// SellerNotificationLog maps to the seller_notification_log table.
type SellerNotificationLog struct {
	ID               int64           `json:"id"`
	SellerID         int64           `json:"seller_id"`
	RecommendationID int64           `json:"recommendation_id"`
	ChannelCode      string          `json:"channel_code"`
	DeliverySystemID *string         `json:"delivery_system_id,omitempty"`
	Status           string          `json:"status"`
	PayloadJSON      *json.RawMessage `json:"payload_json,omitempty"`
	SentAt           *time.Time      `json:"sent_at,omitempty"`
	DeliveredAt      *time.Time      `json:"delivered_at,omitempty"`
	OpenedAt         *time.Time      `json:"opened_at,omitempty"`
	ClickedAt        *time.Time      `json:"clicked_at,omitempty"`
	ErrorMessage     *string         `json:"error_message,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}
