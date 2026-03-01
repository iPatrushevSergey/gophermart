package worker

import (
	"context"
	"time"

	"gophermart/internal/gophermart/application/port"
	modulefactory "gophermart/internal/gophermart/modules/orders/presentation/factory"
)

// Starter describes a background worker that can be started with context.
type Starter interface {
	Start(ctx context.Context)
}

// RegistryParams contains dependencies required to build orders workers.
type RegistryParams struct {
	UseCases     modulefactory.UseCaseFactory
	Log          port.Logger
	PollInterval time.Duration
}

// BuildWorkers builds all orders module background workers.
func BuildWorkers(p RegistryParams) []Starter {
	return []Starter{
		NewAccrualWorker(p.UseCases, p.Log, p.PollInterval),
	}
}
