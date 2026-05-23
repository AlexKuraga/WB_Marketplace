package notifications

import "context"

// HandoffAdapter transfers ready notifications to the downstream delivery service.
type HandoffAdapter struct{}

// New creates a notification handoff adapter.
func New() *HandoffAdapter {
	return &HandoffAdapter{}
}

// Dispatch sends pending outbound notifications. Integration logic will be added later.
func (a *HandoffAdapter) Dispatch(ctx context.Context) error {
	_ = a
	_ = ctx
	return nil
}
