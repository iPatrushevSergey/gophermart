package service

import (
	"time"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// Withdrawal Service creates a write-off domain object.
// Moved to the service (and not the constructor in the entity) so that a single
// point of creation and, if necessary, add rules (limits, duplicates by order_number)
// without smearing the use case.
type WithdrawalService struct{}

// Create creates a debit fact.
func (WithdrawalService) Create(userID vo.UserID, orderNumber vo.OrderNumber, amount vo.Points, at time.Time) entity.Withdrawal {
	return entity.Withdrawal{
		UserID:      userID,
		OrderNumber: orderNumber,
		Amount:      amount,
		ProcessedAt: at,
	}
}
