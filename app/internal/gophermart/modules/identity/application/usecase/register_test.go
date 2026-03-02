package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"gophermart/internal/gophermart/application"
	appmocks "gophermart/internal/gophermart/application/port/mocks"
	"gophermart/internal/gophermart/modules/identity/application/dto"
	identityportmocks "gophermart/internal/gophermart/modules/identity/application/port/mocks"
	"gophermart/internal/gophermart/modules/identity/domain/entity"
	"gophermart/internal/gophermart/modules/identity/domain/vo"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type stubBalanceGateway struct {
	openAccount func(ctx context.Context, userID vo.UserID, createdAt time.Time) error
}

func (s *stubBalanceGateway) OpenAccount(ctx context.Context, userID vo.UserID, createdAt time.Time) error {
	if s.openAccount != nil {
		return s.openAccount(ctx, userID, createdAt)
	}
	return nil
}

func TestRegisterUser_Execute(t *testing.T) {
	ctx := context.Background()
	input := dto.RegisterInput{Login: "alice", Password: "secret"}
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		userReader := identityportmocks.NewMockUserReader(ctrl)
		userWriter := identityportmocks.NewMockUserWriter(ctrl)
		balanceGateway := &stubBalanceGateway{}
		transactor := appmocks.NewMockTransactor(ctrl)
		hasher := appmocks.NewMockPasswordHasher(ctrl)
		clk := appmocks.NewMockClock(ctrl)

		userReader.EXPECT().FindByLogin(ctx, "alice").Return(nil, application.ErrNotFound)
		hasher.EXPECT().Hash("secret").Return("hashed", nil)
		clk.EXPECT().Now().Return(fixedTime)
		transactor.EXPECT().RunInTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		userWriter.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, u *entity.User) error {
				u.ID = vo.UserID(1)
				assert.Equal(t, fixedTime, u.CreatedAt)
				return nil
			},
		)
		balanceGateway.openAccount = func(_ context.Context, userID vo.UserID, createdAt time.Time) error {
			assert.Equal(t, vo.UserID(1), userID)
			assert.Equal(t, fixedTime, createdAt)
			return nil
		}

		uc := NewRegisterUser(userReader, userWriter, balanceGateway, transactor, hasher, clk)
		id, err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		assert.Equal(t, vo.UserID(1), id)
	})

	t.Run("login already taken", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		userReader := identityportmocks.NewMockUserReader(ctrl)
		userReader.EXPECT().FindByLogin(ctx, "alice").Return(&entity.User{Login: "alice"}, nil)

		uc := NewRegisterUser(userReader, nil, nil, nil, nil, nil)
		_, err := uc.Execute(ctx, input)

		assert.ErrorIs(t, err, application.ErrAlreadyExists)
	})

	t.Run("hash error", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		userReader := identityportmocks.NewMockUserReader(ctrl)
		hasher := appmocks.NewMockPasswordHasher(ctrl)

		userReader.EXPECT().FindByLogin(ctx, "alice").Return(nil, application.ErrNotFound)
		hasher.EXPECT().Hash("secret").Return("", errors.New("hash failed"))

		uc := NewRegisterUser(userReader, nil, nil, nil, hasher, nil)
		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "hash failed")
	})

	t.Run("user repo create error", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		userReader := identityportmocks.NewMockUserReader(ctrl)
		userWriter := identityportmocks.NewMockUserWriter(ctrl)
		transactor := appmocks.NewMockTransactor(ctrl)
		hasher := appmocks.NewMockPasswordHasher(ctrl)
		clk := appmocks.NewMockClock(ctrl)

		userReader.EXPECT().FindByLogin(ctx, "alice").Return(nil, application.ErrNotFound)
		hasher.EXPECT().Hash("secret").Return("hashed", nil)
		clk.EXPECT().Now().Return(fixedTime)
		transactor.EXPECT().RunInTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		userWriter.EXPECT().Create(ctx, gomock.Any()).Return(errors.New("db error"))

		uc := NewRegisterUser(userReader, userWriter, nil, transactor, hasher, clk)
		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
	})

	t.Run("find by login unexpected error", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		userReader := identityportmocks.NewMockUserReader(ctrl)
		userReader.EXPECT().FindByLogin(ctx, "alice").Return(nil, errors.New("connection lost"))

		uc := NewRegisterUser(userReader, nil, nil, nil, nil, nil)
		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection lost")
	})
}
