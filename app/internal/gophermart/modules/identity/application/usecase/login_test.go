package usecase

import (
	"context"
	"errors"
	"testing"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port/mocks"
	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestLoginUser_Execute(t *testing.T) {
	ctx := context.Background()
	input := dto.LoginInput{Login: "alice", Password: "secret"}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		userReader := mocks.NewMockUserReader(ctrl)
		hasher := mocks.NewMockPasswordHasher(ctrl)

		userReader.EXPECT().FindByLogin(ctx, "alice").Return(&entity.User{
			ID: vo.UserID(1), Login: "alice", PasswordHash: "hashed",
		}, nil)
		hasher.EXPECT().Compare("secret", "hashed").Return(true)

		uc := NewLoginUser(userReader, hasher)
		id, err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		assert.Equal(t, vo.UserID(1), id)
	})

	t.Run("user not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		userReader := mocks.NewMockUserReader(ctrl)
		userReader.EXPECT().FindByLogin(ctx, "alice").Return(nil, nil)

		uc := NewLoginUser(userReader, nil)
		_, err := uc.Execute(ctx, input)

		assert.ErrorIs(t, err, application.ErrInvalidCredentials)
	})

	t.Run("wrong password", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		userReader := mocks.NewMockUserReader(ctrl)
		hasher := mocks.NewMockPasswordHasher(ctrl)

		userReader.EXPECT().FindByLogin(ctx, "alice").Return(&entity.User{
			ID: vo.UserID(1), PasswordHash: "hashed",
		}, nil)
		hasher.EXPECT().Compare("secret", "hashed").Return(false)

		uc := NewLoginUser(userReader, hasher)
		_, err := uc.Execute(ctx, input)

		assert.ErrorIs(t, err, application.ErrInvalidCredentials)
	})

	t.Run("repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		userReader := mocks.NewMockUserReader(ctrl)
		userReader.EXPECT().FindByLogin(ctx, "alice").Return(nil, errors.New("db error"))

		uc := NewLoginUser(userReader, nil)
		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
	})
}
