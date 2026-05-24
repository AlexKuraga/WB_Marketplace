package rules

import (
	"testing"
	"time"

	"wb-marketplace/internal/domain"
)

func TestEvaluateSnapshot(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		snapshot domain.SellerMetricsSnapshot
		want     []string
	}{
		{
			name: "no products",
			snapshot: domain.SellerMetricsSnapshot{
				ActiveProductsCount: 0,
			},
			want: []string{"NO_PRODUCTS"},
		},
		{
			name: "no sales",
			snapshot: domain.SellerMetricsSnapshot{
				ActiveProductsCount: 5,
				NoSalesDays:         14,
			},
			want: []string{"NO_SALES"},
		},
		{
			name: "inactive seller",
			snapshot: domain.SellerMetricsSnapshot{
				ActiveProductsCount: 1,
				LastLoginDays:       7,
			},
			want: []string{"INACTIVE_SELLER"},
		},
		{
			name: "out of stock",
			snapshot: domain.SellerMetricsSnapshot{
				ActiveProductsCount:       3,
				ProductsWithoutStockCount: 2,
			},
			want: []string{"OUT_OF_STOCK"},
		},
		{
			name: "multiple matches",
			snapshot: domain.SellerMetricsSnapshot{
				ActiveProductsCount:       2,
				NoSalesDays:               20,
				LastLoginDays:             10,
				ProductsWithoutStockCount: 1,
			},
			want: []string{"NO_SALES", "INACTIVE_SELLER", "OUT_OF_STOCK"},
		},
		{
			name: "no matches",
			snapshot: domain.SellerMetricsSnapshot{
				ActiveProductsCount: 2,
				NoSalesDays:         3,
				LastLoginDays:       1,
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := engine.EvaluateSnapshot(tt.snapshot)
			if len(matches) != len(tt.want) {
				t.Fatalf("len(matches) = %d, want %d", len(matches), len(tt.want))
			}

			for i, code := range tt.want {
				if matches[i].TriggerCode != code {
					t.Errorf("matches[%d].TriggerCode = %q, want %q", i, matches[i].TriggerCode, code)
				}
			}
		})
	}
}

func TestEvaluateSnapshotRecommendationFields(t *testing.T) {
	engine := New()
	matches := engine.EvaluateSnapshot(domain.SellerMetricsSnapshot{ActiveProductsCount: 0})
	if len(matches) != 1 {
		t.Fatalf("len(matches) = %d, want 1", len(matches))
	}

	match := matches[0]
	if match.RecommendationTypeID != 1 {
		t.Errorf("RecommendationTypeID = %d, want 1", match.RecommendationTypeID)
	}
	if match.Title == "" || match.Description == "" || match.ReasonText == "" {
		t.Fatal("expected recommendation text fields to be set")
	}
}

func TestEvaluateSnapshotUsesSnapshotDateForPeriodKeySource(t *testing.T) {
	engine := New()
	snapshot := domain.SellerMetricsSnapshot{
		SnapshotDate:        time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		ActiveProductsCount: 0,
	}
	matches := engine.EvaluateSnapshot(snapshot)
	if len(matches) == 0 {
		t.Fatal("expected at least one match")
	}
}
