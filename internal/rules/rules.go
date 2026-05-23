package rules

import "context"

// Engine applies business rules, deduplication, and cooldown checks.
type Engine struct{}

// New creates a rule engine instance.
func New() *Engine {
	return &Engine{}
}

// Evaluate matches seller features against configured rules. Logic will be added later.
func (e *Engine) Evaluate(ctx context.Context) error {
	_ = e
	_ = ctx
	return nil
}
