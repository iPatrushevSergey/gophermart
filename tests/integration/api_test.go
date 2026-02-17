//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophermart/internal/gophermart/adapters/accrual"
	"gophermart/internal/gophermart/adapters/auth"
	adapterclock "gophermart/internal/gophermart/adapters/clock"
	"gophermart/internal/gophermart/adapters/repository/postgres"
	"gophermart/internal/gophermart/adapters/validation"
	portmocks "gophermart/internal/gophermart/application/port/mocks"
	"gophermart/internal/gophermart/domain/service"
	"gophermart/internal/gophermart/presentation/http/handler"
	"gophermart/internal/gophermart/testutil"

	"gophermart/cmd/gophermart/bootstrap"

	"go.uber.org/mock/gomock"
)

// setupE2EServer creates a full application stack with real DB and returns
// an httptest.Server ready for HTTP requests.
func setupE2EServer(t *testing.T) *httptest.Server {
	t.Helper()

	pool := testutil.SetupPostgres(t)

	transactor := postgres.NewTransactor(pool, postgres.RetryConfig{
		MaxRetries: 1,
		BaseDelay:  50 * time.Millisecond,
		MaxDelay:   200 * time.Millisecond,
	})

	hasher := auth.NewBCryptHasher(4) // low cost for fast tests
	tokens := auth.NewJWTProvider("test-secret", 1*time.Hour)
	luhnValidator := validation.NewLuhnValidator()
	clk := adapterclock.Real{}

	// Accrual client is unused in E2E (no background worker), pass a stub.
	accrualHTTP := &http.Client{Timeout: 1 * time.Second}
	accrualClient := accrual.NewClient("http://localhost:1", accrualHTTP)

	userRepo := postgres.NewUserRepository(transactor)
	orderRepo := postgres.NewOrderRepository(transactor)
	balanceRepo := postgres.NewBalanceAccountRepository(transactor)
	withdrawalRepo := postgres.NewWithdrawalRepository(transactor)

	userSvc := service.UserService{}
	balanceSvc := service.BalanceService{}
	orderSvc := service.OrderService{}
	withdrawalSvc := service.WithdrawalService{}

	ctrl := gomock.NewController(t)
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	log.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	log.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	ucFactory := bootstrap.NewUseCaseFactory(
		userRepo, orderRepo, balanceRepo, withdrawalRepo,
		hasher, tokens, transactor, luhnValidator, accrualClient, clk,
		userSvc, balanceSvc, orderSvc, withdrawalSvc, log,
		50, 3,
	)

	userHandler := handler.NewUserHandler(ucFactory, tokens, log)
	orderHandler := handler.NewOrderHandler(ucFactory, log)
	balanceHandler := handler.NewBalanceHandler(ucFactory, log)

	router := bootstrap.NewRouter(userHandler, orderHandler, balanceHandler, tokens, log)

	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	return ts
}

func doJSON(t *testing.T, client *http.Client, method, url string, body any) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, reader)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func doText(t *testing.T, client *http.Client, method, url, text string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, bytes.NewReader([]byte(text)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func extractToken(t *testing.T, resp *http.Response) string {
	t.Helper()
	authHeader := resp.Header.Get("Authorization")
	require.NotEmpty(t, authHeader, "Authorization header must be set")
	require.Contains(t, authHeader, "Bearer ")
	return authHeader[len("Bearer "):]
}

func authedClient(token string) *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &tokenTransport{token: token},
	}
}

type tokenTransport struct {
	token string
}

func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return http.DefaultTransport.RoundTrip(req)
}

// TestE2E_FullUserFlow tests the complete user journey:
// register -> login -> upload order -> list orders -> get balance -> withdraw -> list withdrawals.
func TestE2E_FullUserFlow(t *testing.T) {
	ts := setupE2EServer(t)
	client := &http.Client{}

	// 1. Register a new user.
	resp := doJSON(t, client, http.MethodPost, ts.URL+"/api/user/register",
		map[string]string{"login": "e2e-user", "password": "password123"})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	token := extractToken(t, resp)
	resp.Body.Close()

	ac := authedClient(token)

	// 2. Register the same user again — conflict.
	resp = doJSON(t, client, http.MethodPost, ts.URL+"/api/user/register",
		map[string]string{"login": "e2e-user", "password": "password123"})
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	resp.Body.Close()

	// 3. Login with the registered user.
	resp = doJSON(t, client, http.MethodPost, ts.URL+"/api/user/login",
		map[string]string{"login": "e2e-user", "password": "password123"})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	loginToken := extractToken(t, resp)
	assert.NotEmpty(t, loginToken)
	resp.Body.Close()

	// 4. Login with wrong password.
	resp = doJSON(t, client, http.MethodPost, ts.URL+"/api/user/login",
		map[string]string{"login": "e2e-user", "password": "wrong"})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// 5. Upload a valid order number (passes Luhn check: 12345678903).
	resp = doText(t, ac, http.MethodPost, ts.URL+"/api/user/orders", "12345678903")
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	resp.Body.Close()

	// 6. Upload the same order again — already exists (own).
	resp = doText(t, ac, http.MethodPost, ts.URL+"/api/user/orders", "12345678903")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// 7. Upload an invalid order number (fails Luhn check).
	resp = doText(t, ac, http.MethodPost, ts.URL+"/api/user/orders", "1234567890")
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	resp.Body.Close()

	// 8. List orders — should have one order.
	resp = doJSON(t, ac, http.MethodGet, ts.URL+"/api/user/orders", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var orders []map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&orders))
	resp.Body.Close()
	assert.Len(t, orders, 1)
	assert.Equal(t, "12345678903", orders[0]["number"])
	assert.Equal(t, "NEW", orders[0]["status"])

	// 9. Get balance — should be zero.
	resp = doJSON(t, ac, http.MethodGet, ts.URL+"/api/user/balance", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var balance map[string]float64
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&balance))
	resp.Body.Close()
	assert.Equal(t, float64(0), balance["current"])
	assert.Equal(t, float64(0), balance["withdrawn"])

	// 10. Attempt withdraw with zero balance — insufficient.
	resp = doJSON(t, ac, http.MethodPost, ts.URL+"/api/user/balance/withdraw",
		map[string]any{"order": "12345678903", "sum": 100})
	assert.Equal(t, http.StatusPaymentRequired, resp.StatusCode)
	resp.Body.Close()

	// 11. List withdrawals — empty.
	resp = doJSON(t, ac, http.MethodGet, ts.URL+"/api/user/withdrawals", nil)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// 12. Access protected endpoint without token — unauthorized.
	resp = doJSON(t, client, http.MethodGet, ts.URL+"/api/user/orders", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

// TestE2E_OrderConflictBetweenUsers verifies that an order uploaded by one user
// cannot be uploaded by another user (409 Conflict).
func TestE2E_OrderConflictBetweenUsers(t *testing.T) {
	ts := setupE2EServer(t)
	client := &http.Client{}

	// Register user A.
	resp := doJSON(t, client, http.MethodPost, ts.URL+"/api/user/register",
		map[string]string{"login": "user-a", "password": "pass"})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	tokenA := extractToken(t, resp)
	resp.Body.Close()

	// Register user B.
	resp = doJSON(t, client, http.MethodPost, ts.URL+"/api/user/register",
		map[string]string{"login": "user-b", "password": "pass"})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	tokenB := extractToken(t, resp)
	resp.Body.Close()

	acA := authedClient(tokenA)
	acB := authedClient(tokenB)

	// User A uploads order.
	resp = doText(t, acA, http.MethodPost, ts.URL+"/api/user/orders", "12345678903")
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	resp.Body.Close()

	// User B tries to upload the same order — conflict.
	resp = doText(t, acB, http.MethodPost, ts.URL+"/api/user/orders", "12345678903")
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	resp.Body.Close()
}
