package service

import (
	"context"
	"fmt"
	"strings"

	"wb-marketplace/internal/domain"
)

// CreateRuleInput contains fields for creating a recommendation rule.
type CreateRuleInput struct {
	RuleCode             string
	RuleName             string
	Description          *string
	RecommendationTypeID int64
	Priority             int
	CooldownDays         int
	IsActive             bool
	ConditionExpression  string
	CreatedBy            *string
}

// RuleService implements admin rule use cases.
type RuleService struct {
	repos repositoryBundle
}

// NewRuleService creates a rule service instance.
func NewRuleService(repos repositoryBundle) *RuleService {
	return &RuleService{repos: repos}
}

// ListActive returns active recommendation rules.
func (s *RuleService) ListActive(ctx context.Context) ([]domain.RecommendationRule, error) {
	if s.repos.Rules == nil {
		return []domain.RecommendationRule{}, nil
	}

	rules, err := s.repos.Rules.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	if rules == nil {
		return []domain.RecommendationRule{}, nil
	}
	return rules, nil
}

// Create validates input and creates a new recommendation rule.
func (s *RuleService) Create(ctx context.Context, input CreateRuleInput) (domain.RecommendationRule, error) {
	if err := validateCreateRuleInput(input); err != nil {
		return domain.RecommendationRule{}, err
	}
	if s.repos.Rules == nil {
		return domain.RecommendationRule{}, fmt.Errorf("rule repository not configured")
	}

	rule := domain.RecommendationRule{
		RuleCode:             strings.TrimSpace(input.RuleCode),
		RuleName:             strings.TrimSpace(input.RuleName),
		Description:          input.Description,
		RecommendationTypeID: input.RecommendationTypeID,
		Priority:             input.Priority,
		CooldownDays:         input.CooldownDays,
		IsActive:             input.IsActive,
		ConditionExpression:  strings.TrimSpace(input.ConditionExpression),
		CreatedBy:            input.CreatedBy,
		UpdatedBy:            input.CreatedBy,
	}

	if err := s.repos.Rules.Create(ctx, rule); err != nil {
		return domain.RecommendationRule{}, err
	}

	return rule, nil
}

func validateCreateRuleInput(input CreateRuleInput) error {
	if strings.TrimSpace(input.RuleCode) == "" {
		return fmt.Errorf("rule_code is required")
	}
	if strings.TrimSpace(input.RuleName) == "" {
		return fmt.Errorf("rule_name is required")
	}
	if input.RecommendationTypeID <= 0 {
		return fmt.Errorf("recommendation_type_id must be positive")
	}
	if strings.TrimSpace(input.ConditionExpression) == "" {
		return fmt.Errorf("condition_expression is required")
	}
	if input.Priority < 0 {
		return fmt.Errorf("priority must be non-negative")
	}
	if input.CooldownDays < 0 {
		return fmt.Errorf("cooldown_days must be non-negative")
	}
	return nil
}
