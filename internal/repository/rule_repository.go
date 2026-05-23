package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wb-marketplace/internal/domain"
)

// RuleRepository loads and creates recommendation rules.
type RuleRepository interface {
	ListActive(ctx context.Context) ([]domain.RecommendationRule, error)
	Create(ctx context.Context, rule domain.RecommendationRule) error
}

type postgresRuleRepository struct {
	pool *pgxpool.Pool
}

// NewRuleRepository creates a PostgreSQL-backed rule repository.
func NewRuleRepository(pool *pgxpool.Pool) RuleRepository {
	return &postgresRuleRepository{pool: pool}
}

const selectRecommendationRuleColumns = `
	id, rule_code, rule_name, description, recommendation_type_id,
	priority, cooldown_days, is_active, condition_expression,
	created_by, updated_by, created_at, updated_at
`

func (r *postgresRuleRepository) ListActive(ctx context.Context) ([]domain.RecommendationRule, error) {
	query := `
		SELECT ` + selectRecommendationRuleColumns + `
		FROM recommendation_rules
		WHERE is_active = TRUE
		ORDER BY priority ASC, id ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query active rules: %w", err)
	}
	defer rows.Close()

	rules := make([]domain.RecommendationRule, 0)
	for rows.Next() {
		item, err := scanRecommendationRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active rules: %w", err)
	}

	return rules, nil
}

func (r *postgresRuleRepository) Create(ctx context.Context, rule domain.RecommendationRule) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO recommendation_rules (
			rule_code, rule_name, description, recommendation_type_id,
			priority, cooldown_days, is_active, condition_expression,
			created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		rule.RuleCode,
		rule.RuleName,
		rule.Description,
		rule.RecommendationTypeID,
		rule.Priority,
		rule.CooldownDays,
		rule.IsActive,
		rule.ConditionExpression,
		rule.CreatedBy,
		rule.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert recommendation rule: %w", err)
	}
	return nil
}

func scanRecommendationRule(row pgx.Row) (domain.RecommendationRule, error) {
	var item domain.RecommendationRule
	err := row.Scan(
		&item.ID,
		&item.RuleCode,
		&item.RuleName,
		&item.Description,
		&item.RecommendationTypeID,
		&item.Priority,
		&item.CooldownDays,
		&item.IsActive,
		&item.ConditionExpression,
		&item.CreatedBy,
		&item.UpdatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return domain.RecommendationRule{}, fmt.Errorf("scan recommendation rule: %w", err)
	}
	return item, nil
}
