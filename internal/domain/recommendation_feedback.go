package domain

import (
	"encoding/json"
	"time"
)

// RecommendationFeedback maps to the recommendation_feedback table.
type RecommendationFeedback struct {
	ID               int64           `json:"id"`
	SellerID         int64           `json:"seller_id"`
	RecommendationID int64           `json:"recommendation_id"`
	FeedbackType     string          `json:"feedback_type"`
	FeedbackAt       time.Time       `json:"feedback_at"`
	PayloadJSON      *json.RawMessage `json:"payload_json,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}
