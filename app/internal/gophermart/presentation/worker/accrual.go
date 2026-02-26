package worker

import (
	"context"
	"errors"
	"time"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/presentation/factory"
)

// AccrualWorker polls the accrual system and updates order statuses.
type AccrualWorker struct {
	processAccrual port.BackgroundRunner
	log            port.Logger
	pollInterval   time.Duration
}

// NewAccrualWorker creates a new accrual background worker.
func NewAccrualWorker(ucFactory factory.UseCaseFactory, log port.Logger, pollInterval time.Duration) *AccrualWorker {
	return &AccrualWorker{
		processAccrual: ucFactory.ProcessAccrualUseCase(),
		log:            log,
		pollInterval:   pollInterval,
	}
}

// Start runs the worker loop in a goroutine. Cancel ctx to stop.
func (w *AccrualWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

func (w *AccrualWorker) run(ctx context.Context) {
	w.log.Info("accrual worker started", "poll_interval", w.pollInterval)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("accrual worker stopped")
			return
		case <-ticker.C:
			w.poll(ctx)
		}
	}
}

func (w *AccrualWorker) poll(ctx context.Context) {
	processed, err := w.processAccrual.Run(ctx)
	if err != nil {
		var rateLimit *application.ErrRateLimit
		if errors.As(err, &rateLimit) {
			w.log.Warn("accrual rate limited, backing off",
				"retry_after", rateLimit.RetryAfter,
			)
			select {
			case <-ctx.Done():
			case <-time.After(rateLimit.RetryAfter):
			}
			return
		}
		w.log.Error("accrual poll failed", "error", err)
		return
	}

	if processed > 0 {
		w.log.Debug("accrual batch processed", "count", processed)
	}
}
