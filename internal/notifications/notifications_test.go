package notifications

import (
	"context"
	"testing"
)

func TestDispatchNoOp(t *testing.T) {
	adapter := New()
	if err := adapter.Dispatch(context.Background()); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
}
