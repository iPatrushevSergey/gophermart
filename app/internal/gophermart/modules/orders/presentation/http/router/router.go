package router

import (
	"github.com/gin-gonic/gin"

	"gophermart/internal/gophermart/application/port"
	modulefactory "gophermart/internal/gophermart/modules/orders/presentation/factory"
	ordershandler "gophermart/internal/gophermart/modules/orders/presentation/http/handler"
)

// RegisterProtectedRoutes registers protected orders endpoints.
func RegisterProtectedRoutes(
	protected *gin.RouterGroup,
	useCases modulefactory.UseCaseFactory,
	log port.Logger,
) {
	orderHandler := ordershandler.NewOrderHandler(useCases, log)
	protected.POST("/orders", orderHandler.Upload)
	protected.GET("/orders", orderHandler.List)
}
