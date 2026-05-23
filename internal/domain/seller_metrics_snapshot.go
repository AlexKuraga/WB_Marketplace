package domain

import "time"

// SellerMetricsSnapshot maps to the seller_metrics_snapshot table.
type SellerMetricsSnapshot struct {
	ID                         int64     `json:"id"`
	SellerID                   int64     `json:"seller_id"`
	SnapshotDate               time.Time `json:"snapshot_date"`
	ActiveProductsCount        int       `json:"active_products_count"`
	PublishedProductsCount     int       `json:"published_products_count"`
	ProductsWithoutStockCount  int       `json:"products_without_stock_count"`
	CategoriesCount            int       `json:"categories_count"`
	ActiveCategoriesCount      int       `json:"active_categories_count"`
	RegionsCount               int       `json:"regions_count"`
	Orders7d                   int       `json:"orders_7d"`
	Orders30d                  int       `json:"orders_30d"`
	Revenue7d                  float64   `json:"revenue_7d"`
	Revenue30d                 float64   `json:"revenue_30d"`
	Margin30d                  float64   `json:"margin_30d"`
	LastLoginDays              int       `json:"last_login_days"`
	NoSalesDays                int       `json:"no_sales_days"`
	CurrentPrimaryModelCode    *string   `json:"current_primary_model_code,omitempty"`
	CreatedAt                  time.Time `json:"created_at"`
}
