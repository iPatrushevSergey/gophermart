package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/dto"
	portmocks "gophermart/internal/gophermart/application/port/mocks"
	"gophermart/internal/gophermart/domain/vo"
	"gophermart/internal/gophermart/presentation/factory/mocks"
	"gophermart/internal/gophermart/presentation/http/handler"
	"gophermart/internal/gophermart/presentation/http/httpcontext"
)

func setupBalanceRouter(t *testing.T) (*gomock.Controller, *mocks.MockUseCaseFactory, *gin.Engine) {
	t.Helper()
	ctrl := gomock.NewController(t)
	factory := mocks.NewMockUseCaseFactory(ctrl)
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	h := handler.NewBalanceHandler(factory, log)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	authSim := func(userID vo.UserID) gin.HandlerFunc {
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

	factory.EXPECT().GetBalanceUseCase().Return(&stubUseCase[vo.UserID, dto.BalanceOutput]{
		out: dto.BalanceOutput{Current: 500.5, Withdrawn: 42},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]float64
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 500.5, resp["current"])
	assert.Equal(t, float64(42), resp["withdrawn"])
}

func TestBalanceHandler_Withdraw_Success(t *testing.T) {
	_, factory, router := setupBalanceRouter(t)

	factory.EXPECT().WithdrawUseCase().Return(&stubUseCase[dto.WithdrawInput, struct{}]{})

	body, _ := json.Marshal(map[string]any{"order": "12345678903", "sum": 100.5})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBalanceHandler_Withdraw_InsufficientBalance(t *testing.T) {
	_, factory, router := setupBalanceRouter(t)

	factory.EXPECT().WithdrawUseCase().Return(&stubUseCase[dto.WithdrawInput, struct{}]{err: application.ErrInsufficientBalance})

	body, _ := json.Marshal(map[string]any{"order": "12345678903", "sum": 9999})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusPaymentRequired, w.Code)
}

func TestBalanceHandler_Withdraw_InvalidOrder(t *testing.T) {
	_, factory, router := setupBalanceRouter(t)

	factory.EXPECT().WithdrawUseCase().Return(&stubUseCase[dto.WithdrawInput, struct{}]{err: application.ErrInvalidOrderNumber})

	body, _ := json.Marshal(map[string]any{"order": "bad", "sum": 10})
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
	factory.EXPECT().ListWithdrawalsUseCase().Return(&stubUseCase[vo.UserID, []dto.WithdrawalOutput]{out: withdrawals})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, "12345678903", resp[0]["order"])
}

func TestBalanceHandler_ListWithdrawals_Empty(t *testing.T) {
	_, factory, router := setupBalanceRouter(t)

	factory.EXPECT().ListWithdrawalsUseCase().Return(&stubUseCase[vo.UserID, []dto.WithdrawalOutput]{out: []dto.WithdrawalOutput{}})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
