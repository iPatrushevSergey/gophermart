package intermodule

import (
	"context"
	"time"

	"gophermart/internal/gophermart/domain/vo"
	balanceapi "gophermart/internal/gophermart/modules/balance/application/api"
)

// BalanceGatewayAdapter bridges orders module to balance module API.
type BalanceGatewayAdapter struct {
	api balanceapi.AccrualAPI
}

func NewBalanceGatewayAdapter(api balanceapi.AccrualAPI) *BalanceGatewayAdapter {
	return &BalanceGatewayAdapter{api: api}
}

func (a *BalanceGatewayAdapter) ApplyAccrual(ctx context.Context, userID vo.UserID, amount vo.Points, processedAt time.Time) error {
	return a.api.ApplyAccrual(ctx, balanceapi.ApplyAccrualInput{
		UserID:      userID,
		Amount:      amount,
		ProcessedAt: processedAt,
	})
}
