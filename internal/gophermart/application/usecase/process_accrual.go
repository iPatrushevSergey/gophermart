package usecase

import (
	"context"
	"errors"

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
		optimisticRetries: optimisticRetries,
	}
}

// Run processes one batch of pending orders. Returns the number of processed orders.
func (uc *ProcessAccrual) Run(ctx context.Context) (int, error) {
	orders, err := uc.orderReader.ListByStatuses(ctx, []entity.OrderStatus{
		entity.OrderStatusNew,
		entity.OrderStatusProcessing,
	}, uc.batchSize)
	if err != nil {
		return 0, err
	}

	processed := 0
	for _, order := range orders {
		if ctx.Err() != nil {
			break
		}

		err := uc.processOrder(ctx, order)
		if err != nil {
			var rl *application.ErrRateLimit
			if errors.As(err, &rl) {
				return processed, err
			}
			uc.log.Warn("failed to process order accrual",
				"order", order.Number.String(),
				"error", err,
			)
			continue
		}
		processed++
	}

	return processed, nil
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
