package contract_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	balanceapi "gophermart/internal/gophermart/modules/balance/application/api"
	ordersintermodule "gophermart/internal/gophermart/modules/orders/adapters/intermodule"
	ordersport "gophermart/internal/gophermart/modules/orders/application/port"
	ordersvo "gophermart/internal/gophermart/modules/orders/domain/vo"
)

type accrualAPISpy struct {
	called bool
	gotCtx context.Context
	gotIn  balanceapi.ApplyAccrualInput
	err    error
}

func (s *accrualAPISpy) ApplyAccrual(ctx context.Context, in balanceapi.ApplyAccrualInput) error {
	s.called = true
	s.gotCtx = ctx
	s.gotIn = in
	return s.err
}

func TestOrdersBalanceGatewayContract(t *testing.T) {
	// Compile-time contract: orders adapter must satisfy consumer-owned port.
	var _ ordersport.BalanceGateway = (*ordersintermodule.BalanceGatewayAdapter)(nil)

	ctx := context.Background()
	now := time.Date(2026, 1, 1, 13, 0, 0, 0, time.UTC)
	userID := ordersvo.UserID(101)
	amount := ordersvo.Points(123.45)

	t.Run("maps input to provider API", func(t *testing.T) {
		api := &accrualAPISpy{}
		adapter := ordersintermodule.NewBalanceGatewayAdapter(api)

		err := adapter.ApplyAccrual(ctx, userID, amount, now)
		require.NoError(t, err)
		require.True(t, api.called)
		assert.Equal(t, ctx, api.gotCtx)
		assert.Equal(t, balanceapi.ApplyAccrualInput{
			UserID:      int64(userID),
			Amount:      float64(amount),
			ProcessedAt: now,
		}, api.gotIn)
	})

	t.Run("propagates provider API error", func(t *testing.T) {
		expectedErr := errors.New("provider failed")
		api := &accrualAPISpy{err: expectedErr}
		adapter := ordersintermodule.NewBalanceGatewayAdapter(api)

		err := adapter.ApplyAccrual(ctx, userID, amount, now)
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}
