package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wb-marketplace/internal/domain"
)

// AnalysisRepository supports batch analysis reads and writes.
type AnalysisRepository interface {
	ListSellersForAnalysis(ctx context.Context) ([]domain.Seller, error)
	GetLatestSnapshot(ctx context.Context, sellerID int64) (*domain.SellerMetricsSnapshot, error)
	CreateTrigger(ctx context.Context, trigger domain.SellerTriggerLog) error
	CreateRecommendation(ctx context.Context, recommendation domain.Recommendation) error
	CreateNotificationLog(ctx context.Context, notification domain.SellerNotificationLog) error
	CreateAnalysisJob(ctx context.Context, job domain.AnalysisJob) (int64, error)
	UpdateAnalysisJob(ctx context.Context, job domain.AnalysisJob) error
}

type postgresAnalysisRepository struct {
	pool *pgxpool.Pool
}

// NewAnalysisRepository creates a PostgreSQL-backed analysis repository.
func NewAnalysisRepository(pool *pgxpool.Pool) AnalysisRepository {
	return &postgresAnalysisRepository{pool: pool}
}

const selectSellerColumns = `
	id, external_seller_id, seller_name, seller_type, status, lifecycle_stage,
	registration_at, last_login_at, home_region_id, created_at, updated_at
`

const selectSellerMetricsSnapshotColumns = `
	id, seller_id, snapshot_date, active_products_count, published_products_count,
	products_without_stock_count, categories_count, active_categories_count,
	regions_count, orders_7d, orders_30d, revenue_7d, revenue_30d, margin_30d,
	last_login_days, no_sales_days, current_primary_model_code, created_at
`

func (r *postgresAnalysisRepository) ListSellersForAnalysis(ctx context.Context) ([]domain.Seller, error) {
	query := `
		SELECT ` + selectSellerColumns + `
		FROM sellers
		WHERE status = 'active'
		ORDER BY id ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query sellers for analysis: %w", err)
	}
	defer rows.Close()

	sellers := make([]domain.Seller, 0)
	for rows.Next() {
		item, err := scanSeller(rows)
		if err != nil {
			return nil, err
		}
		sellers = append(sellers, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sellers for analysis: %w", err)
	}

	return sellers, nil
}

func (r *postgresAnalysisRepository) GetLatestSnapshot(ctx context.Context, sellerID int64) (*domain.SellerMetricsSnapshot, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+selectSellerMetricsSnapshotColumns+`
		FROM seller_metrics_snapshot
		WHERE seller_id = $1
		ORDER BY snapshot_date DESC, id DESC
		LIMIT 1
	`, sellerID)

	snapshot, err := scanSellerMetricsSnapshot(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &snapshot, nil
}

func (r *postgresAnalysisRepository) CreateTrigger(ctx context.Context, trigger domain.SellerTriggerLog) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO seller_trigger_log (
			seller_id, rule_id, trigger_code, triggered_at, period_key,
			snapshot_id, payload_json, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		trigger.SellerID,
		trigger.RuleID,
		trigger.TriggerCode,
		trigger.TriggeredAt,
		trigger.PeriodKey,
		trigger.SnapshotID,
		trigger.PayloadJSON,
		trigger.Status,
	)
	if err != nil {
		return fmt.Errorf("insert seller trigger log: %w", err)
	}
	return nil
}

func (r *postgresAnalysisRepository) CreateRecommendation(ctx context.Context, recommendation domain.Recommendation) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO recommendations (
			seller_id, trigger_id, recommendation_type_id, template_id,
			title, description, reason_text, priority, score, status, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		recommendation.SellerID,
		recommendation.TriggerID,
		recommendation.RecommendationTypeID,
		recommendation.TemplateID,
		recommendation.Title,
		recommendation.Description,
		recommendation.ReasonText,
		recommendation.Priority,
		recommendation.Score,
		recommendation.Status,
		recommendation.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert recommendation: %w", err)
	}
	return nil
}

func (r *postgresAnalysisRepository) CreateNotificationLog(ctx context.Context, notification domain.SellerNotificationLog) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO seller_notification_log (
			seller_id, recommendation_id, channel_code, delivery_system_id,
			status, payload_json, sent_at, delivered_at, opened_at, clicked_at, error_message
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		notification.SellerID,
		notification.RecommendationID,
		notification.ChannelCode,
		notification.DeliverySystemID,
		notification.Status,
		notification.PayloadJSON,
		notification.SentAt,
		notification.DeliveredAt,
		notification.OpenedAt,
		notification.ClickedAt,
		notification.ErrorMessage,
	)
	if err != nil {
		return fmt.Errorf("insert seller notification log: %w", err)
	}
	return nil
}

func (r *postgresAnalysisRepository) CreateAnalysisJob(ctx context.Context, job domain.AnalysisJob) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO analysis_jobs (
			job_type, status, started_at, finished_at,
			sellers_processed, recommendations_created, triggers_created, error_message
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`,
		job.JobType,
		job.Status,
		job.StartedAt,
		job.FinishedAt,
		job.SellersProcessed,
		job.RecommendationsCreated,
		job.TriggersCreated,
		job.ErrorMessage,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert analysis job: %w", err)
	}
	return id, nil
}

func (r *postgresAnalysisRepository) UpdateAnalysisJob(ctx context.Context, job domain.AnalysisJob) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE analysis_jobs
		SET status = $2,
		    finished_at = $3,
		    sellers_processed = $4,
		    recommendations_created = $5,
		    triggers_created = $6,
		    error_message = $7
		WHERE id = $1
	`,
		job.ID,
		job.Status,
		job.FinishedAt,
		job.SellersProcessed,
		job.RecommendationsCreated,
		job.TriggersCreated,
		job.ErrorMessage,
	)
	if err != nil {
		return fmt.Errorf("update analysis job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("analysis job %d not found", job.ID)
	}
	return nil
}

func scanSeller(row pgx.Row) (domain.Seller, error) {
	var item domain.Seller
	err := row.Scan(
		&item.ID,
		&item.ExternalSellerID,
		&item.SellerName,
		&item.SellerType,
		&item.Status,
		&item.LifecycleStage,
		&item.RegistrationAt,
		&item.LastLoginAt,
		&item.HomeRegionID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return domain.Seller{}, fmt.Errorf("scan seller: %w", err)
	}
	return item, nil
}

func scanSellerMetricsSnapshot(row pgx.Row) (domain.SellerMetricsSnapshot, error) {
	var item domain.SellerMetricsSnapshot
	err := row.Scan(
		&item.ID,
		&item.SellerID,
		&item.SnapshotDate,
		&item.ActiveProductsCount,
		&item.PublishedProductsCount,
		&item.ProductsWithoutStockCount,
		&item.CategoriesCount,
		&item.ActiveCategoriesCount,
		&item.RegionsCount,
		&item.Orders7d,
		&item.Orders30d,
		&item.Revenue7d,
		&item.Revenue30d,
		&item.Margin30d,
		&item.LastLoginDays,
		&item.NoSalesDays,
		&item.CurrentPrimaryModelCode,
		&item.CreatedAt,
	)
	if err != nil {
		return domain.SellerMetricsSnapshot{}, fmt.Errorf("scan seller metrics snapshot: %w", err)
	}
	return item, nil
}
