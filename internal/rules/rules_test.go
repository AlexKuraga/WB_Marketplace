package rules

import (
	"context"
	"testing"
)

func TestEvaluateNoOp(t *testing.T) {
	engine := New()
	if err := engine.Evaluate(context.Background()); err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
}
