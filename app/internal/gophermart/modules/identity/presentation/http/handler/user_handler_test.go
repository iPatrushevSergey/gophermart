package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	portmocks "gophermart/internal/gophermart/application/port/mocks"
	"gophermart/internal/gophermart/domain/vo"
	identityhandler "gophermart/internal/gophermart/modules/identity/presentation/http/handler"
	"gophermart/internal/gophermart/presentation/http/httpcontext"
)

// stubUseCase is a simple implementation of port.UseCase[In, Out] for tests.
type stubUseCase[In, Out any] struct {
	out Out
	err error
}

func (s *stubUseCase[In, Out]) Execute(_ context.Context, _ In) (Out, error) {
	return s.out, s.err
}

type testIdentityFactory struct {
	registerUC port.UseCase[dto.RegisterInput, vo.UserID]
	loginUC    port.UseCase[dto.LoginInput, vo.UserID]
}

func (f *testIdentityFactory) RegisterUseCase() port.UseCase[dto.RegisterInput, vo.UserID] {
	return f.registerUC
}

func (f *testIdentityFactory) LoginUseCase() port.UseCase[dto.LoginInput, vo.UserID] {
	return f.loginUC
}

func setupUserRouter(t *testing.T) (*gomock.Controller, *testIdentityFactory, *portmocks.MockTokenProvider, *gin.Engine) {
	t.Helper()
	ctrl := gomock.NewController(t)
	factory := &testIdentityFactory{}
	tokens := portmocks.NewMockTokenProvider(ctrl)
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	h := identityhandler.NewUserHandler(factory, tokens, log)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/user/register", h.Register)
	r.POST("/api/user/login", h.Login)

	return ctrl, factory, tokens, r
}

func TestUserHandler_Register_Success(t *testing.T) {
	_, factory, tokens, router := setupUserRouter(t)

	var userID vo.UserID = 42
	factory.registerUC = &stubUseCase[dto.RegisterInput, vo.UserID]{out: userID}
	tokens.EXPECT().Issue(userID).Return("test-jwt-token", nil)

	body, err := json.Marshal(map[string]string{"login": "alice", "password": "secret123"})
	require.NoError(t, err)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Authorization"), "Bearer test-jwt-token")

	result := w.Result()
	defer result.Body.Close()
	cookies := result.Cookies()
	require.NotEmpty(t, cookies)
	var found bool
	for _, c := range cookies {
		if c.Name == httpcontext.CookieName {
			assert.Equal(t, "test-jwt-token", c.Value)
			found = true
		}
	}
	assert.True(t, found, "auth cookie not set")
}

func TestUserHandler_Register_AlreadyExists(t *testing.T) {
	_, factory, _, router := setupUserRouter(t)

	factory.registerUC = &stubUseCase[dto.RegisterInput, vo.UserID]{err: application.ErrAlreadyExists}

	body, err := json.Marshal(map[string]string{"login": "alice", "password": "secret123"})
	require.NoError(t, err)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUserHandler_Register_BadJSON(t *testing.T) {
	_, _, _, router := setupUserRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Login_Success(t *testing.T) {
	_, factory, tokens, router := setupUserRouter(t)

	var userID vo.UserID = 7
	factory.loginUC = &stubUseCase[dto.LoginInput, vo.UserID]{out: userID}
	tokens.EXPECT().Issue(userID).Return("login-token", nil)

	body, err := json.Marshal(map[string]string{"login": "alice", "password": "secret123"})
	require.NoError(t, err)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Authorization"), "Bearer login-token")
}

func TestUserHandler_Login_InvalidCredentials(t *testing.T) {
	_, factory, _, router := setupUserRouter(t)

	factory.loginUC = &stubUseCase[dto.LoginInput, vo.UserID]{err: application.ErrInvalidCredentials}

	body, err := json.Marshal(map[string]string{"login": "alice", "password": "wrong"})
	require.NoError(t, err)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserHandler_Login_BadJSON(t *testing.T) {
	_, _, _, router := setupUserRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
