package worker

import "context"

// Starter describes a background worker that can be started with context.
type Starter interface {
	Start(ctx context.Context)
}

// RegistryParams contains dependencies required to build identity workers.
type RegistryParams struct{}

// BuildWorkers builds all identity module background workers.
func BuildWorkers(_ RegistryParams) []Starter {
	return []Starter{}
}
