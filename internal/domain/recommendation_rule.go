package domain

import "time"

// RecommendationRule maps to the recommendation_rules table.
type RecommendationRule struct {
	ID                   int64     `json:"id"`
	RuleCode             string    `json:"rule_code"`
	RuleName             string    `json:"rule_name"`
	Description          *string   `json:"description,omitempty"`
	RecommendationTypeID int64     `json:"recommendation_type_id"`
	Priority             int       `json:"priority"`
	CooldownDays         int       `json:"cooldown_days"`
	IsActive             bool      `json:"is_active"`
	ConditionExpression  string    `json:"condition_expression"`
	CreatedBy            *string   `json:"created_by,omitempty"`
	UpdatedBy            *string   `json:"updated_by,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
