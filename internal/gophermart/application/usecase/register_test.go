package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port/mocks"
	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/service"
	"gophermart/internal/gophermart/domain/vo"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegisterUser_Execute(t *testing.T) {
	ctx := context.Background()
	input := dto.RegisterInput{Login: "alice", Password: "secret"}
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		userReader := mocks.NewMockUserReader(ctrl)
		userWriter := mocks.NewMockUserWriter(ctrl)
		balanceWriter := mocks.NewMockBalanceAccountWriter(ctrl)
		transactor := mocks.NewMockTransactor(ctrl)
		hasher := mocks.NewMockPasswordHasher(ctrl)
		clk := mocks.NewMockClock(ctrl)

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
		balanceWriter.EXPECT().Create(ctx, gomock.Any()).Return(nil)

		uc := NewRegisterUser(userReader, userWriter, balanceWriter, transactor, hasher, clk, service.BalanceService{})
		id, err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		assert.Equal(t, vo.UserID(1), id)
	})

	t.Run("login already taken", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		userReader := mocks.NewMockUserReader(ctrl)
		userReader.EXPECT().FindByLogin(ctx, "alice").Return(&entity.User{Login: "alice"}, nil)

		uc := NewRegisterUser(userReader, nil, nil, nil, nil, nil, service.BalanceService{})
		_, err := uc.Execute(ctx, input)

		assert.ErrorIs(t, err, application.ErrAlreadyExists)
	})

	t.Run("hash error", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		userReader := mocks.NewMockUserReader(ctrl)
		hasher := mocks.NewMockPasswordHasher(ctrl)

		userReader.EXPECT().FindByLogin(ctx, "alice").Return(nil, application.ErrNotFound)
		hasher.EXPECT().Hash("secret").Return("", errors.New("hash failed"))

		uc := NewRegisterUser(userReader, nil, nil, nil, hasher, nil, service.BalanceService{})
		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "hash failed")
	})

	t.Run("user repo create error", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		userReader := mocks.NewMockUserReader(ctrl)
		userWriter := mocks.NewMockUserWriter(ctrl)
		transactor := mocks.NewMockTransactor(ctrl)
		hasher := mocks.NewMockPasswordHasher(ctrl)
		clk := mocks.NewMockClock(ctrl)

		userReader.EXPECT().FindByLogin(ctx, "alice").Return(nil, application.ErrNotFound)
		hasher.EXPECT().Hash("secret").Return("hashed", nil)
		clk.EXPECT().Now().Return(fixedTime)
		transactor.EXPECT().RunInTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		userWriter.EXPECT().Create(ctx, gomock.Any()).Return(errors.New("db error"))

		uc := NewRegisterUser(userReader, userWriter, nil, transactor, hasher, clk, service.BalanceService{})
		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
	})

	t.Run("find by login unexpected error", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		userReader := mocks.NewMockUserReader(ctrl)
		userReader.EXPECT().FindByLogin(ctx, "alice").Return(nil, errors.New("connection lost"))

		uc := NewRegisterUser(userReader, nil, nil, nil, nil, nil, service.BalanceService{})
		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection lost")
	})
}
