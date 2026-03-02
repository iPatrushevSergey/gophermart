package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/port"
	portmocks "gophermart/internal/gophermart/application/port/mocks"
	"gophermart/internal/gophermart/modules/balance/application/dto"
	"gophermart/internal/gophermart/modules/balance/domain/vo"
	"gophermart/internal/gophermart/modules/balance/presentation/http/handler"
	"gophermart/internal/gophermart/presentation/http/httpcontext"
)

type stubUseCase[In, Out any] struct {
	out Out
	err error
}

func (s *stubUseCase[In, Out]) Execute(_ context.Context, _ In) (Out, error) {
	return s.out, s.err
}

type testBalanceFactory struct {
	getBalanceUC      port.UseCase[vo.UserID, dto.BalanceOutput]
	withdrawUC        port.UseCase[dto.WithdrawInput, struct{}]
	listWithdrawalsUC port.UseCase[vo.UserID, []dto.WithdrawalOutput]
}

func (f *testBalanceFactory) GetBalanceUseCase() port.UseCase[vo.UserID, dto.BalanceOutput] {
	return f.getBalanceUC
}

func (f *testBalanceFactory) WithdrawUseCase() port.UseCase[dto.WithdrawInput, struct{}] {
	return f.withdrawUC
}

func (f *testBalanceFactory) ListWithdrawalsUseCase() port.UseCase[vo.UserID, []dto.WithdrawalOutput] {
	return f.listWithdrawalsUC
}

func setupBalanceRouter(t *testing.T) (*gomock.Controller, *testBalanceFactory, *gin.Engine) {
	t.Helper()
	ctrl := gomock.NewController(t)
	factory := &testBalanceFactory{}
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	h := handler.NewBalanceHandler(factory, log)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	authSim := func(userID int64) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set(httpcontext.UserIDKey, userID)
			c.Next()
		}
	}

	protected := r.Group("", authSim(1))
	protected.GET("/api/user/balance", h.Get)
	protected.POST("/api/user/balance/withdraw", h.Withdraw)
	protected.GET("/api/user/withdrawals", h.ListWithdrawals)

	return ctrl, factory, r
}

func TestBalanceHandler_Get_Success(t *testing.T) {
	_, factory, router := setupBalanceRouter(t)

	factory.getBalanceUC = &stubUseCase[vo.UserID, dto.BalanceOutput]{
		out: dto.BalanceOutput{Current: 500.5, Withdrawn: 42},
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]float64
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 500.5, resp["current"])
	assert.Equal(t, float64(42), resp["withdrawn"])
}

func TestBalanceHandler_Withdraw_Success(t *testing.T) {
	_, factory, router := setupBalanceRouter(t)
	factory.withdrawUC = &stubUseCase[dto.WithdrawInput, struct{}]{}

	body, err := json.Marshal(map[string]any{"order": "12345678903", "sum": 100.5})
	require.NoError(t, err)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBalanceHandler_Withdraw_InsufficientBalance(t *testing.T) {
	_, factory, router := setupBalanceRouter(t)
	factory.withdrawUC = &stubUseCase[dto.WithdrawInput, struct{}]{err: application.ErrInsufficientBalance}

	body, err := json.Marshal(map[string]any{"order": "12345678903", "sum": 9999})
	require.NoError(t, err)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusPaymentRequired, w.Code)
}

func TestBalanceHandler_Withdraw_InvalidOrder(t *testing.T) {
	_, factory, router := setupBalanceRouter(t)
	factory.withdrawUC = &stubUseCase[dto.WithdrawInput, struct{}]{err: application.ErrInvalidOrderNumber}

	body, err := json.Marshal(map[string]any{"order": "bad", "sum": 10})
	require.NoError(t, err)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestBalanceHandler_Withdraw_BadJSON(t *testing.T) {
	_, _, router := setupBalanceRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader([]byte("broken")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBalanceHandler_ListWithdrawals_Success(t *testing.T) {
	_, factory, router := setupBalanceRouter(t)

	withdrawals := []dto.WithdrawalOutput{
		{OrderNumber: "12345678903", Sum: 100, ProcessedAt: time.Date(2026, 1, 20, 12, 0, 0, 0, time.UTC)},
		{OrderNumber: "99999999927", Sum: 50, ProcessedAt: time.Date(2026, 1, 21, 8, 0, 0, 0, time.UTC)},
	}
	factory.listWithdrawalsUC = &stubUseCase[vo.UserID, []dto.WithdrawalOutput]{out: withdrawals}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, "12345678903", resp[0]["order"])
}

func TestBalanceHandler_ListWithdrawals_Empty(t *testing.T) {
	_, factory, router := setupBalanceRouter(t)
	factory.listWithdrawalsUC = &stubUseCase[vo.UserID, []dto.WithdrawalOutput]{out: []dto.WithdrawalOutput{}}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
