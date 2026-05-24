package service

import (
	"context"
	"testing"

	"wb-marketplace/internal/domain"
)

type mockRuleRepository struct {
	listActiveFn func(ctx context.Context) ([]domain.RecommendationRule, error)
	createFn       func(ctx context.Context, rule domain.RecommendationRule) error
}

func (m *mockRuleRepository) ListActive(ctx context.Context) ([]domain.RecommendationRule, error) {
	return m.listActiveFn(ctx)
}

func (m *mockRuleRepository) Create(ctx context.Context, rule domain.RecommendationRule) error {
	return m.createFn(ctx, rule)
}

func TestRuleServiceListActiveReturnsEmptySlice(t *testing.T) {
	svc := NewRuleService(repositoryBundle{
		Rules: &mockRuleRepository{
			listActiveFn: func(ctx context.Context) ([]domain.RecommendationRule, error) {
				return nil, nil
			},
		},
	})

	rules, err := svc.ListActive(context.Background())
	if err != nil {
		t.Fatalf("ListActive() error = %v", err)
	}
	if rules == nil {
		t.Fatal("expected empty slice, got nil")
	}
}

func TestRuleServiceCreateValidation(t *testing.T) {
	svc := NewRuleService(repositoryBundle{
		Rules: &mockRuleRepository{},
	})

	_, err := svc.Create(context.Background(), CreateRuleInput{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRuleServiceCreate(t *testing.T) {
	var created domain.RecommendationRule

	svc := NewRuleService(repositoryBundle{
		Rules: &mockRuleRepository{
			createFn: func(ctx context.Context, rule domain.RecommendationRule) error {
				created = rule
				return nil
			},
		},
	})

	input := CreateRuleInput{
		RuleCode:             "NO_PRODUCTS",
		RuleName:             "No products",
		RecommendationTypeID: 1,
		Priority:             10,
		CooldownDays:         7,
		IsActive:             true,
		ConditionExpression:  "active_products_count = 0",
	}

	rule, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if rule.RuleCode != input.RuleCode {
		t.Errorf("RuleCode = %q, want %q", rule.RuleCode, input.RuleCode)
	}
	if created.RuleCode != input.RuleCode {
		t.Errorf("created RuleCode = %q, want %q", created.RuleCode, input.RuleCode)
	}
}
