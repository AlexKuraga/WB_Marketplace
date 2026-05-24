package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wb-marketplace/internal/domain"
)

// RecommendationRepository loads and updates seller recommendations and feedback.
type RecommendationRepository interface {
	GetActiveBySellerID(ctx context.Context, sellerID int64) ([]domain.Recommendation, error)
	UpdateStatus(ctx context.Context, recommendationID int64, status string) error
	CreateFeedback(ctx context.Context, recommendationID int64, feedbackType string) error
}

type postgresRecommendationRepository struct {
	pool *pgxpool.Pool
}

// NewRecommendationRepository creates a PostgreSQL-backed recommendation repository.
func NewRecommendationRepository(pool *pgxpool.Pool) RecommendationRepository {
	return &postgresRecommendationRepository{pool: pool}
}

const selectRecommendationColumns = `
	id, seller_id, trigger_id, recommendation_type_id, template_id,
	title, description, reason_text, priority, score, status,
	expires_at, created_at, updated_at
`

func (r *postgresRecommendationRepository) GetActiveBySellerID(ctx context.Context, sellerID int64) ([]domain.Recommendation, error) {
	query := `
		SELECT ` + selectRecommendationColumns + `
		FROM recommendations
		WHERE seller_id = $1
		  AND status IN ('created', 'ready_to_send', 'sent', 'opened')
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY priority ASC, created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, sellerID)
	if err != nil {
		return nil, fmt.Errorf("query active recommendations: %w", err)
	}
	defer rows.Close()

	recommendations := make([]domain.Recommendation, 0)
	for rows.Next() {
		item, err := scanRecommendation(rows)
		if err != nil {
			return nil, err
		}
		recommendations = append(recommendations, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active recommendations: %w", err)
	}

	return recommendations, nil
}

func (r *postgresRecommendationRepository) UpdateStatus(ctx context.Context, recommendationID int64, status string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE recommendations
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`, recommendationID, status)
	if err != nil {
		return fmt.Errorf("update recommendation status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("recommendation %d not found", recommendationID)
	}
	return nil
}

func (r *postgresRecommendationRepository) CreateFeedback(ctx context.Context, recommendationID int64, feedbackType string) error {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO recommendation_feedback (seller_id, recommendation_id, feedback_type)
		SELECT seller_id, $1, $2
		FROM recommendations
		WHERE id = $1
	`, recommendationID, feedbackType)
	if err != nil {
		return fmt.Errorf("insert recommendation feedback: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("recommendation %d not found", recommendationID)
	}
	return nil
}

func scanRecommendation(row pgx.Row) (domain.Recommendation, error) {
	var item domain.Recommendation
	err := row.Scan(
		&item.ID,
		&item.SellerID,
		&item.TriggerID,
		&item.RecommendationTypeID,
		&item.TemplateID,
		&item.Title,
		&item.Description,
		&item.ReasonText,
		&item.Priority,
		&item.Score,
		&item.Status,
		&item.ExpiresAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return domain.Recommendation{}, fmt.Errorf("scan recommendation: %w", err)
	}
	return item, nil
}
