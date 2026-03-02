package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/modules/orders/application/dto"
	"gophermart/internal/gophermart/modules/orders/domain/vo"
	"gophermart/internal/gophermart/modules/orders/presentation/factory"
	httpdto "gophermart/internal/gophermart/modules/orders/presentation/http/dto"
	"gophermart/internal/gophermart/presentation/http/httpcontext"
)

// maxOrderNumberBytes is a safety limit for order number body size.
const maxOrderNumberBytes = 64

// OrderHandler manages order-related requests.
type OrderHandler struct {
	useCases factory.UseCaseFactory
	log      port.Logger
}

// NewOrderHandler creates an OrderHandler with orders use cases provider.
func NewOrderHandler(useCases factory.UseCaseFactory, log port.Logger) *OrderHandler {
	return &OrderHandler{
		useCases: useCases,
		log:      log,
	}
}

// Upload accepts an order number for accrual calculation.
func (h *OrderHandler) Upload(c *gin.Context) {
	userID, ok := httpcontext.UserID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxOrderNumberBytes))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "empty order number"})
		return
	}

	_, err = h.useCases.UploadOrderUseCase().Execute(
		c.Request.Context(),
		dto.UploadOrderInput{UserID: vo.UserID(userID), OrderNumber: orderNumber},
	)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrAlreadyExists):
			c.Status(http.StatusOK)
		case errors.Is(err, application.ErrConflict):
			c.AbortWithStatus(http.StatusConflict)
		case errors.Is(err, application.ErrInvalidOrderNumber):
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid order number"})
		default:
			h.log.Error("upload order failed", "error", err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	c.Status(http.StatusAccepted)
}

// List returns all orders uploaded by the authenticated user.
func (h *OrderHandler) List(c *gin.Context) {
	userID, ok := httpcontext.UserID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	orders, err := h.useCases.ListOrdersUseCase().Execute(c.Request.Context(), vo.UserID(userID))
	if err != nil {
		h.log.Error("list orders failed", "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	resp := make([]httpdto.OrderResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, httpdto.OrderResponse{
			Number:     o.Number,
			Status:     o.Status,
			Accrual:    o.Accrual,
			UploadedAt: o.UploadedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, resp)
}
