package analysis

import (
	"context"
	"testing"
)

func TestRunNoOp(t *testing.T) {
	svc := New()
	if err := svc.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}
