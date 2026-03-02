package contract_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	balanceapi "gophermart/internal/gophermart/modules/balance/application/api"
	identityintermodule "gophermart/internal/gophermart/modules/identity/adapters/intermodule"
	identityport "gophermart/internal/gophermart/modules/identity/application/port"
	identityvo "gophermart/internal/gophermart/modules/identity/domain/vo"
)

type accountAPISpy struct {
	called bool
	gotCtx context.Context
	gotIn  balanceapi.OpenAccountInput
	err    error
}

func (s *accountAPISpy) OpenAccount(ctx context.Context, in balanceapi.OpenAccountInput) error {
	s.called = true
	s.gotCtx = ctx
	s.gotIn = in
	return s.err
}

func TestIdentityBalanceGatewayContract(t *testing.T) {
	// Compile-time contract: identity adapter must satisfy consumer-owned port.
	var _ identityport.BalanceGateway = (*identityintermodule.BalanceGatewayAdapter)(nil)

	ctx := context.Background()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	userID := identityvo.UserID(42)

	t.Run("maps input to provider API", func(t *testing.T) {
		api := &accountAPISpy{}
		adapter := identityintermodule.NewBalanceGatewayAdapter(api)

		err := adapter.OpenAccount(ctx, userID, now)
		require.NoError(t, err)
		require.True(t, api.called)
		assert.Equal(t, ctx, api.gotCtx)
		assert.Equal(t, balanceapi.OpenAccountInput{
			UserID:    int64(userID),
			CreatedAt: now,
		}, api.gotIn)
	})

	t.Run("propagates provider API error", func(t *testing.T) {
		expectedErr := errors.New("provider failed")
		api := &accountAPISpy{err: expectedErr}
		adapter := identityintermodule.NewBalanceGatewayAdapter(api)

		err := adapter.OpenAccount(ctx, userID, now)
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}
