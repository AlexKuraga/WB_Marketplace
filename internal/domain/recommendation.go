package domain

import "time"

// Recommendation maps to the recommendations table.
type Recommendation struct {
	ID                   int64      `json:"id"`
	SellerID             int64      `json:"seller_id"`
	TriggerID            int64      `json:"trigger_id"`
	RecommendationTypeID int64      `json:"recommendation_type_id"`
	TemplateID           *int64     `json:"template_id,omitempty"`
	Title                string     `json:"title"`
	Description          string     `json:"description"`
	ReasonText           *string    `json:"reason_text,omitempty"`
	Priority             int        `json:"priority"`
	Score                *float64   `json:"score,omitempty"`
	Status               string     `json:"status"`
	ExpiresAt            *time.Time `json:"expires_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
