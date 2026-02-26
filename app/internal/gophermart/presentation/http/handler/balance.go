package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/presentation/factory"
	httpdto "gophermart/internal/gophermart/presentation/http/dto"
	"gophermart/internal/gophermart/presentation/http/httpcontext"
)

// BalanceHandler manages balance and withdrawal requests.
type BalanceHandler struct {
	factory factory.UseCaseFactory
	log     port.Logger
}

// NewBalanceHandler creates a BalanceHandler.
func NewBalanceHandler(factory factory.UseCaseFactory, log port.Logger) *BalanceHandler {
	return &BalanceHandler{factory: factory, log: log}
}

// Get returns the current loyalty balance of the authenticated user.
func (h *BalanceHandler) Get(c *gin.Context) {
	userID, ok := httpcontext.UserID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	balance, err := h.factory.GetBalanceUseCase().Execute(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("get balance failed", "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, httpdto.BalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	})
}

// Withdraw deducts loyalty points from the user's balance.
func (h *BalanceHandler) Withdraw(c *gin.Context) {
	userID, ok := httpcontext.UserID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var req httpdto.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	_, err := h.factory.WithdrawUseCase().Execute(
		c.Request.Context(),
		dto.WithdrawInput{UserID: userID, OrderNumber: req.Order, Sum: req.Sum},
	)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInsufficientBalance):
			c.AbortWithStatus(http.StatusPaymentRequired)
		case errors.Is(err, application.ErrInvalidOrderNumber):
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid order number"})
		default:
			h.log.Error("withdraw failed", "error", err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	c.Status(http.StatusOK)
}

// ListWithdrawals returns the withdrawal history of the authenticated user.
func (h *BalanceHandler) ListWithdrawals(c *gin.Context) {
	userID, ok := httpcontext.UserID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.factory.ListWithdrawalsUseCase().Execute(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("list withdrawals failed", "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	resp := make([]httpdto.WithdrawalResponse, 0, len(withdrawals))
	for _, w := range withdrawals {
		resp = append(resp, httpdto.WithdrawalResponse{
			Order:       w.OrderNumber,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, resp)
}
