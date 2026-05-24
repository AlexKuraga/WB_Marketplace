package service

import (
	"context"
	"testing"
)

func TestRecommendationServiceWithoutRepository(t *testing.T) {
	svc := NewRecommendationService(repositoryBundle{})

	recs, err := svc.GetActiveBySeller(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetActiveBySeller() error = %v", err)
	}

	if recs == nil {
		t.Fatalf("expected empty slice, got nil")
	}

	if len(recs) != 0 {
		t.Fatalf("expected empty slice, got %#v", recs)
	}
}