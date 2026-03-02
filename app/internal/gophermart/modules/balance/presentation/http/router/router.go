package router

import (
	"github.com/gin-gonic/gin"

	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/modules/balance/presentation/factory"
	"gophermart/internal/gophermart/modules/balance/presentation/http/handler"
)

// RegisterProtectedRoutes registers protected balance endpoints.
func RegisterProtectedRoutes(
	protected *gin.RouterGroup,
	useCases factory.UseCaseFactory,
	log port.Logger,
) {
	balanceHandler := handler.NewBalanceHandler(useCases, log)
	protected.GET("/balance", balanceHandler.Get)
	protected.POST("/balance/withdraw", balanceHandler.Withdraw)
	protected.GET("/withdrawals", balanceHandler.ListWithdrawals)
}
