package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/dto"
)

// defaultRetryAfter is a fallback when 429 response has no Retry-After header.
const defaultRetryAfter = 60 * time.Second

// Client is an HTTP implementation of port.AccrualClient.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new accrual HTTP client.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// NewClientFromConfig creates a new accrual client from adapter config.
func NewClientFromConfig(cfg Config) *Client {
	return NewClient(cfg.Address, &http.Client{Timeout: cfg.HTTPTimeout})
}

type accrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

// GetOrderAccrual calls the accrual system.
func (c *Client) GetOrderAccrual(ctx context.Context, orderNumber string) (*dto.AccrualOrderInfo, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("accrual: create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("accrual: do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var body accrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return nil, fmt.Errorf("accrual: decode response: %w", err)
		}
		return &dto.AccrualOrderInfo{
			Status:  body.Status,
			Accrual: body.Accrual,
		}, nil

	case http.StatusNoContent:
		// Order not registered in accrual system
		return nil, nil

	case http.StatusTooManyRequests:
		retryAfter := defaultRetryAfter
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if seconds, err := strconv.Atoi(ra); err == nil {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}
		return nil, &application.ErrRateLimit{RetryAfter: retryAfter}

	default:
		return nil, fmt.Errorf("accrual: unexpected status %d", resp.StatusCode)
	}
}
