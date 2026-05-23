package domain

import "time"

// Seller maps to the sellers table.
type Seller struct {
	ID               int64      `json:"id"`
	ExternalSellerID string     `json:"external_seller_id"`
	SellerName       string     `json:"seller_name"`
	SellerType       string     `json:"seller_type"`
	Status           string     `json:"status"`
	LifecycleStage   string     `json:"lifecycle_stage"`
	RegistrationAt   *time.Time `json:"registration_at,omitempty"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
	HomeRegionID     *int64     `json:"home_region_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
