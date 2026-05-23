package db

import (
	"context"
	"strings"
	"testing"
)

func TestConnectInvalidURL(t *testing.T) {
	ctx := context.Background()

	_, err := Connect(ctx, "not-a-valid-dsn")
	if err == nil {
		t.Fatal("Connect() expected error for invalid DSN")
	}
	if !strings.Contains(err.Error(), "create pool") {
		t.Errorf("error = %v, want create pool failure", err)
	}
}
