package domain

import (
	"encoding/json"
	"time"
)

// SellerTriggerLog maps to the seller_trigger_log table.
type SellerTriggerLog struct {
	ID          int64           `json:"id"`
	SellerID    int64           `json:"seller_id"`
	RuleID      int64           `json:"rule_id"`
	TriggerCode string          `json:"trigger_code"`
	TriggeredAt time.Time       `json:"triggered_at"`
	PeriodKey   string          `json:"period_key"`
	SnapshotID  *int64          `json:"snapshot_id,omitempty"`
	PayloadJSON *json.RawMessage `json:"payload_json,omitempty"`
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
}
