package usecase

import (
	"context"
	"errors"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// ProcessAccrual fetches pending orders and synchronizes their status with the accrual system.
type ProcessAccrual struct {
	orderReader       port.OrderReader
	orderWriter       port.OrderWriter
	balanceReader     port.BalanceAccountReader
	balanceWriter     port.BalanceAccountWriter
	accrualClient     port.AccrualClient
	transactor        port.Transactor
	clock             port.Clock
	log               port.Logger
	batchSize         int
	maxWorkers        int
	optimisticRetries int
}

// NewProcessAccrual returns the process accrual use case.
func NewProcessAccrual(
	orderReader port.OrderReader,
	orderWriter port.OrderWriter,
	balanceReader port.BalanceAccountReader,
	balanceWriter port.BalanceAccountWriter,
	accrualClient port.AccrualClient,
	transactor port.Transactor,
	clock port.Clock,
	log port.Logger,
	batchSize int,
	maxWorkers int,
	optimisticRetries int,
) *ProcessAccrual {
	return &ProcessAccrual{
		orderReader:       orderReader,
		orderWriter:       orderWriter,
		balanceReader:     balanceReader,
		balanceWriter:     balanceWriter,
		accrualClient:     accrualClient,
		transactor:        transactor,
		clock:             clock,
		log:               log,
		batchSize:         batchSize,
		maxWorkers:        maxWorkers,
		optimisticRetries: optimisticRetries,
	}
}

// Run streams a batch of pending orders from the DB and processes them
// concurrently via errgroup. Returns the number of successfully processed orders.
func (uc *ProcessAccrual) Run(ctx context.Context) (int, error) {
	orders := uc.orderReader.StreamByStatuses(ctx, []entity.OrderStatus{
		entity.OrderStatusNew,
		entity.OrderStatusProcessing,
	}, uc.batchSize)

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(uc.maxWorkers)

	ch := make(chan entity.Order, uc.maxWorkers)
	var streamErr error

	go func() {
		defer close(ch)
		for order, err := range orders {
			if err != nil {
				streamErr = err
				return
			}
			select {
			case ch <- order:
			case <-gCtx.Done():
				return
			}
		}
	}()

	var processed atomic.Int32

	for order := range ch {
		g.Go(func() error {
			if err := uc.processOrder(gCtx, order); err != nil {
				var rl *application.ErrRateLimit
				if errors.As(err, &rl) {
					return err
				}
				uc.log.Warn("failed to process order accrual",
					"order", order.Number.String(),
					"error", err,
				)
				return nil
			}
			processed.Add(1)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return int(processed.Load()), err
	}

	if streamErr != nil {
		return int(processed.Load()), streamErr
	}

	return int(processed.Load()), nil
}

func (uc *ProcessAccrual) processOrder(ctx context.Context, order entity.Order) error {
	info, err := uc.accrualClient.GetOrderAccrual(ctx, order.Number.String())
	if err != nil {
		return err
	}

	if info == nil {
		// Order not registered in accrual system yet — skip.
		// TODO: add max attempts or TTL to avoid polling stale orders forever.
		return nil
	}

	now := uc.clock.Now()

	switch info.Status {
	case "PROCESSING", "REGISTERED":
		order.MarkProcessing()
		return uc.orderWriter.Update(ctx, &order)

	case "INVALID":
		order.MarkInvalid(now)
		return uc.orderWriter.Update(ctx, &order)

	case "PROCESSED":
		accrual := vo.Points(0)
		if info.Accrual != nil {
			accrual = vo.Points(*info.Accrual)
		}

		return application.WithOptimisticRetry(uc.optimisticRetries, func() error {
			return uc.transactor.RunInTransaction(ctx, func(ctx context.Context) error {
				order.MarkProcessed(accrual, now)
				if err := uc.orderWriter.Update(ctx, &order); err != nil {
					return err
				}

				if info.Accrual == nil {
					return nil
				}

				acc, err := uc.balanceReader.FindByUserID(ctx, order.UserID)
				if err != nil {
					return err
				}

				acc.AddAccrual(accrual, now)
				return uc.balanceWriter.Update(ctx, acc)
			})
		})
	}

	return nil
}
