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

func setupOrderRouter(t *testing.T) (*gomock.Controller, *mocks.MockUseCaseFactory, *gin.Engine) {
	t.Helper()
	ctrl := gomock.NewController(t)
	factory := mocks.NewMockUseCaseFactory(ctrl)
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	h := handler.NewOrderHandler(factory, log)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Simulates auth middleware: sets user_id in context.
	authSim := func(userID vo.UserID) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set(httpcontext.UserIDKey, userID)
			c.Next()
		}
	}

	protected := r.Group("", authSim(1))
	protected.POST("/api/user/orders", h.Upload)
	protected.GET("/api/user/orders", h.List)

	// No-auth route for unauthorized test.
	r.POST("/api/user/orders/noauth", h.Upload)

	return ctrl, factory, r
}

func TestOrderHandler_Upload_Accepted(t *testing.T) {
	_, factory, router := setupOrderRouter(t)

	factory.EXPECT().UploadOrderUseCase().Return(&stubUseCase[dto.UploadOrderInput, struct{}]{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("12345678903")))
	req.Header.Set("Content-Type", "text/plain")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestOrderHandler_Upload_AlreadyOwn(t *testing.T) {
	_, factory, router := setupOrderRouter(t)

	factory.EXPECT().UploadOrderUseCase().Return(&stubUseCase[dto.UploadOrderInput, struct{}]{err: application.ErrAlreadyExists})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("12345678903")))
	req.Header.Set("Content-Type", "text/plain")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOrderHandler_Upload_Conflict(t *testing.T) {
	_, factory, router := setupOrderRouter(t)

	factory.EXPECT().UploadOrderUseCase().Return(&stubUseCase[dto.UploadOrderInput, struct{}]{err: application.ErrConflict})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("12345678903")))
	req.Header.Set("Content-Type", "text/plain")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestOrderHandler_Upload_InvalidNumber(t *testing.T) {
	_, factory, router := setupOrderRouter(t)

	factory.EXPECT().UploadOrderUseCase().Return(&stubUseCase[dto.UploadOrderInput, struct{}]{err: application.ErrInvalidOrderNumber})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("bad-number")))
	req.Header.Set("Content-Type", "text/plain")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestOrderHandler_Upload_Unauthorized(t *testing.T) {
	_, _, router := setupOrderRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders/noauth", bytes.NewReader([]byte("12345678903")))
	req.Header.Set("Content-Type", "text/plain")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOrderHandler_List_Success(t *testing.T) {
	_, factory, router := setupOrderRouter(t)

	accrual := 500.5
	orders := []dto.OrderOutput{
		{Number: "12345678903", Status: "PROCESSED", Accrual: &accrual, UploadedAt: time.Date(2026, 1, 20, 12, 0, 0, 0, time.UTC)},
		{Number: "99999999927", Status: "NEW", UploadedAt: time.Date(2026, 1, 21, 8, 0, 0, 0, time.UTC)},
	}
	factory.EXPECT().ListOrdersUseCase().Return(&stubUseCase[vo.UserID, []dto.OrderOutput]{out: orders})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, "12345678903", resp[0]["number"])
	assert.Equal(t, "PROCESSED", resp[0]["status"])
}

func TestOrderHandler_List_Empty(t *testing.T) {
	_, factory, router := setupOrderRouter(t)

	factory.EXPECT().ListOrdersUseCase().Return(&stubUseCase[vo.UserID, []dto.OrderOutput]{out: []dto.OrderOutput{}})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
