//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophermart/internal/gophermart/adapters/repository/postgres"
	"gophermart/internal/gophermart/application"
	balancerepopostgres "gophermart/internal/gophermart/modules/balance/adapters/repository/postgres"
	balanceentity "gophermart/internal/gophermart/modules/balance/domain/entity"
	balancevo "gophermart/internal/gophermart/modules/balance/domain/vo"
	identityrepopostgres "gophermart/internal/gophermart/modules/identity/adapters/repository/postgres"
	identityentity "gophermart/internal/gophermart/modules/identity/domain/entity"
	ordersrepopostgres "gophermart/internal/gophermart/modules/orders/adapters/repository/postgres"
	ordersentity "gophermart/internal/gophermart/modules/orders/domain/entity"
	ordersvo "gophermart/internal/gophermart/modules/orders/domain/vo"
	"gophermart/internal/pkg/testutil"
)

func setupTransactor(t *testing.T) *postgres.Transactor {
	t.Helper()
	pool := testutil.SetupPostgres(t)
	return postgres.NewTransactor(pool, postgres.WithMaxRetries(0))
}

// createTestUser inserts a user and returns it with ID populated.
func createTestUser(t *testing.T, repo *identityrepopostgres.UserRepository, login string, now time.Time) *identityentity.User {
	t.Helper()
	u := &identityentity.User{
		Login:        login,
		PasswordHash: "$2a$10$dummyhash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	err := repo.Create(context.Background(), u)
	require.NoError(t, err)
	require.NotZero(t, u.ID)
	return u
}

// --- UserRepository ---

func TestUserRepository_CreateAndFindByID(t *testing.T) {
	tx := setupTransactor(t)
	repo := identityrepopostgres.NewUserRepository(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)

	u := createTestUser(t, repo, "alice", now)

	found, err := repo.FindByID(context.Background(), u.ID)
	require.NoError(t, err)
	assert.Equal(t, u.ID, found.ID)
	assert.Equal(t, "alice", found.Login)
	assert.Equal(t, "$2a$10$dummyhash", found.PasswordHash)
}

func TestUserRepository_FindByLogin(t *testing.T) {
	tx := setupTransactor(t)
	repo := identityrepopostgres.NewUserRepository(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)

	createTestUser(t, repo, "bob", now)

	found, err := repo.FindByLogin(context.Background(), "bob")
	require.NoError(t, err)
	assert.Equal(t, "bob", found.Login)
}

func TestUserRepository_FindByLogin_NotFound(t *testing.T) {
	tx := setupTransactor(t)
	repo := identityrepopostgres.NewUserRepository(tx)

	_, err := repo.FindByLogin(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, application.ErrNotFound)
}

func TestUserRepository_CreateDuplicate(t *testing.T) {
	tx := setupTransactor(t)
	repo := identityrepopostgres.NewUserRepository(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)

	createTestUser(t, repo, "duplicate", now)

	u2 := &identityentity.User{Login: "duplicate", PasswordHash: "hash2", CreatedAt: now, UpdatedAt: now}
	err := repo.Create(context.Background(), u2)
	assert.ErrorIs(t, err, application.ErrAlreadyExists)
}

// --- OrderRepository ---

func TestOrderRepository_CreateAndFindByNumber(t *testing.T) {
	tx := setupTransactor(t)
	userRepo := identityrepopostgres.NewUserRepository(tx)
	orderRepo := ordersrepopostgres.NewOrderRepository(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := createTestUser(t, userRepo, "order-user", now)

	order := &ordersentity.Order{
		Number:     ordersvo.OrderNumber("12345678903"),
		UserID:     ordersvo.UserID(user.ID),
		Status:     ordersentity.OrderStatusNew,
		UploadedAt: now,
	}
	err := orderRepo.Create(context.Background(), order)
	require.NoError(t, err)

	found, err := orderRepo.FindByNumber(context.Background(), ordersvo.OrderNumber("12345678903"))
	require.NoError(t, err)
	assert.Equal(t, ordersvo.OrderNumber("12345678903"), found.Number)
	assert.Equal(t, ordersvo.UserID(user.ID), found.UserID)
	assert.Equal(t, ordersentity.OrderStatusNew, found.Status)
}

func TestOrderRepository_ListByUserID(t *testing.T) {
	tx := setupTransactor(t)
	userRepo := identityrepopostgres.NewUserRepository(tx)
	orderRepo := ordersrepopostgres.NewOrderRepository(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := createTestUser(t, userRepo, "list-user", now)

	for i, num := range []string{"11111111111", "22222222222"} {
		o := &ordersentity.Order{
			Number:     ordersvo.OrderNumber(num),
			UserID:     ordersvo.UserID(user.ID),
			Status:     ordersentity.OrderStatusNew,
			UploadedAt: now.Add(time.Duration(i) * time.Second),
		}
		require.NoError(t, orderRepo.Create(context.Background(), o))
	}

	orders, err := orderRepo.ListByUserID(context.Background(), ordersvo.UserID(user.ID))
	require.NoError(t, err)
	assert.Len(t, orders, 2)
	// Sorted by uploaded_at DESC, so the second inserted order comes first.
	assert.Equal(t, ordersvo.OrderNumber("22222222222"), orders[0].Number)
}

func TestOrderRepository_ListByStatuses(t *testing.T) {
	tx := setupTransactor(t)
	userRepo := identityrepopostgres.NewUserRepository(tx)
	orderRepo := ordersrepopostgres.NewOrderRepository(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := createTestUser(t, userRepo, "status-user", now)

	newOrder := &ordersentity.Order{Number: "33333333333", UserID: ordersvo.UserID(user.ID), Status: ordersentity.OrderStatusNew, UploadedAt: now}
	require.NoError(t, orderRepo.Create(context.Background(), newOrder))

	procOrder := &ordersentity.Order{Number: "44444444444", UserID: ordersvo.UserID(user.ID), Status: ordersentity.OrderStatusProcessed, UploadedAt: now, Accrual: ptrFloat(100)}
	require.NoError(t, orderRepo.Create(context.Background(), procOrder))

	orders, err := orderRepo.ListByStatuses(context.Background(), []ordersentity.OrderStatus{ordersentity.OrderStatusNew}, 10)
	require.NoError(t, err)
	assert.Len(t, orders, 1)
	assert.Equal(t, ordersentity.OrderStatusNew, orders[0].Status)
}

func TestOrderRepository_Update(t *testing.T) {
	tx := setupTransactor(t)
	userRepo := identityrepopostgres.NewUserRepository(tx)
	orderRepo := ordersrepopostgres.NewOrderRepository(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := createTestUser(t, userRepo, "update-user", now)

	o := &ordersentity.Order{Number: "55555555555", UserID: ordersvo.UserID(user.ID), Status: ordersentity.OrderStatusNew, UploadedAt: now}
	require.NoError(t, orderRepo.Create(context.Background(), o))

	processedAt := now.Add(time.Minute)
	accrual := ordersvo.Points(250.5)
	o.MarkProcessed(accrual, processedAt)
	require.NoError(t, orderRepo.Update(context.Background(), o))

	found, err := orderRepo.FindByNumber(context.Background(), o.Number)
	require.NoError(t, err)
	assert.Equal(t, ordersentity.OrderStatusProcessed, found.Status)
	require.NotNil(t, found.Accrual)
	assert.InDelta(t, 250.5, float64(*found.Accrual), 0.01)
}

func TestOrderRepository_CreateDuplicate(t *testing.T) {
	tx := setupTransactor(t)
	userRepo := identityrepopostgres.NewUserRepository(tx)
	orderRepo := ordersrepopostgres.NewOrderRepository(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := createTestUser(t, userRepo, "dup-order-user", now)

	o1 := &ordersentity.Order{Number: "66666666666", UserID: ordersvo.UserID(user.ID), Status: ordersentity.OrderStatusNew, UploadedAt: now}
	require.NoError(t, orderRepo.Create(context.Background(), o1))

	o2 := &ordersentity.Order{Number: "66666666666", UserID: ordersvo.UserID(user.ID), Status: ordersentity.OrderStatusNew, UploadedAt: now}
	err := orderRepo.Create(context.Background(), o2)
	assert.ErrorIs(t, err, application.ErrAlreadyExists)
}

// --- BalanceAccountRepository ---

func TestBalanceAccountRepository_CreateAndFindByUserID(t *testing.T) {
	tx := setupTransactor(t)
	userRepo := identityrepopostgres.NewUserRepository(tx)
	balanceRepo := balancerepopostgres.NewBalanceAccountRepository(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := createTestUser(t, userRepo, "balance-user", now)

	acc := &balanceentity.BalanceAccount{
		UserID:    balancevo.UserID(user.ID),
		Current:   100.5,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   0,
	}
	require.NoError(t, balanceRepo.Create(context.Background(), acc))

	found, err := balanceRepo.FindByUserID(context.Background(), balancevo.UserID(user.ID))
	require.NoError(t, err)
	assert.Equal(t, balancevo.UserID(user.ID), found.UserID)
	assert.InDelta(t, 100.5, float64(found.Current), 0.01)
	assert.Equal(t, int64(0), found.Version)
}

func TestBalanceAccountRepository_Update(t *testing.T) {
	tx := setupTransactor(t)
	userRepo := identityrepopostgres.NewUserRepository(tx)
	balanceRepo := balancerepopostgres.NewBalanceAccountRepository(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := createTestUser(t, userRepo, "upd-balance-user", now)

	acc := &balanceentity.BalanceAccount{UserID: balancevo.UserID(user.ID), Current: 200, CreatedAt: now, UpdatedAt: now, Version: 0}
	require.NoError(t, balanceRepo.Create(context.Background(), acc))

	acc.Current = 150
	acc.WithdrawnTotal = 50
	acc.UpdatedAt = now.Add(time.Second)
	require.NoError(t, balanceRepo.Update(context.Background(), acc))
	assert.Equal(t, int64(1), acc.Version)

	found, err := balanceRepo.FindByUserID(context.Background(), balancevo.UserID(user.ID))
	require.NoError(t, err)
	assert.InDelta(t, 150, float64(found.Current), 0.01)
	assert.InDelta(t, 50, float64(found.WithdrawnTotal), 0.01)
	assert.Equal(t, int64(1), found.Version)
}

func TestBalanceAccountRepository_OptimisticLock(t *testing.T) {
	tx := setupTransactor(t)
	userRepo := identityrepopostgres.NewUserRepository(tx)
	balanceRepo := balancerepopostgres.NewBalanceAccountRepository(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := createTestUser(t, userRepo, "lock-user", now)

	acc := &balanceentity.BalanceAccount{UserID: balancevo.UserID(user.ID), Current: 300, CreatedAt: now, UpdatedAt: now, Version: 0}
	require.NoError(t, balanceRepo.Create(context.Background(), acc))

	// First update succeeds.
	acc.Current = 250
	acc.UpdatedAt = now.Add(time.Second)
	require.NoError(t, balanceRepo.Update(context.Background(), acc))

	// Simulate stale version: set version back to 0.
	stale := *acc
	stale.Version = 0
	stale.Current = 999
	stale.UpdatedAt = now.Add(2 * time.Second)
	err := balanceRepo.Update(context.Background(), &stale)
	assert.ErrorIs(t, err, application.ErrOptimisticLock)
}

// --- WithdrawalRepository ---

func TestWithdrawalRepository_CreateAndListByUserID(t *testing.T) {
	tx := setupTransactor(t)
	userRepo := identityrepopostgres.NewUserRepository(tx)
	withdrawalRepo := balancerepopostgres.NewWithdrawalRepository(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := createTestUser(t, userRepo, "withdrawal-user", now)

	w1 := &balanceentity.Withdrawal{
		UserID:      balancevo.UserID(user.ID),
		OrderNumber: balancevo.OrderNumber("77777777777"),
		Amount:      50,
		ProcessedAt: now,
	}
	w2 := &balanceentity.Withdrawal{
		UserID:      balancevo.UserID(user.ID),
		OrderNumber: balancevo.OrderNumber("88888888888"),
		Amount:      30,
		ProcessedAt: now.Add(time.Minute),
	}
	require.NoError(t, withdrawalRepo.Create(context.Background(), w1))
	require.NoError(t, withdrawalRepo.Create(context.Background(), w2))

	list, err := withdrawalRepo.ListByUserID(context.Background(), balancevo.UserID(user.ID))
	require.NoError(t, err)
	assert.Len(t, list, 2)
	// Sorted by processed_at DESC, so w2 comes first.
	assert.Equal(t, balancevo.OrderNumber("88888888888"), list[0].OrderNumber)
	assert.Equal(t, balancevo.OrderNumber("77777777777"), list[1].OrderNumber)
}

func TestWithdrawalRepository_ListByUserID_Empty(t *testing.T) {
	tx := setupTransactor(t)
	userRepo := identityrepopostgres.NewUserRepository(tx)
	withdrawalRepo := balancerepopostgres.NewWithdrawalRepository(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)

	user := createTestUser(t, userRepo, "empty-wd-user", now)

	list, err := withdrawalRepo.ListByUserID(context.Background(), balancevo.UserID(user.ID))
	require.NoError(t, err)
	assert.Empty(t, list)
}

func ptrFloat(v float64) *ordersvo.Points {
	p := ordersvo.Points(v)
	return &p
}
