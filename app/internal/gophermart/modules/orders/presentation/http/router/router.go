package router

import (
	"github.com/gin-gonic/gin"

	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/modules/orders/presentation/factory"
	"gophermart/internal/gophermart/modules/orders/presentation/http/handler"
)

// RegisterProtectedRoutes registers protected orders endpoints.
func RegisterProtectedRoutes(
	protected *gin.RouterGroup,
	useCases factory.UseCaseFactory,
	log port.Logger,
) {
	orderHandler := handler.NewOrderHandler(useCases, log)
	protected.POST("/orders", orderHandler.Upload)
	protected.GET("/orders", orderHandler.List)
}
