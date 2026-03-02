package intermodule

import (
	"context"
	"time"

	balanceapi "gophermart/internal/gophermart/modules/balance/application/api"
	"gophermart/internal/gophermart/modules/identity/domain/vo"
)

// BalanceGatewayAdapter bridges identity module to balance module account API.
type BalanceGatewayAdapter struct {
	api balanceapi.AccountAPI
}

func NewBalanceGatewayAdapter(api balanceapi.AccountAPI) *BalanceGatewayAdapter {
	return &BalanceGatewayAdapter{api: api}
}

func (a *BalanceGatewayAdapter) OpenAccount(ctx context.Context, userID vo.UserID, createdAt time.Time) error {
	return a.api.OpenAccount(ctx, balanceapi.OpenAccountInput{
		UserID:    int64(userID),
		CreatedAt: createdAt,
	})
}
