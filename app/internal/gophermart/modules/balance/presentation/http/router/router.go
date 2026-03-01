package router

import (
	"github.com/gin-gonic/gin"

	"gophermart/internal/gophermart/application/port"
	modulefactory "gophermart/internal/gophermart/modules/balance/presentation/factory"
	balancehandler "gophermart/internal/gophermart/modules/balance/presentation/http/handler"
)

// RegisterProtectedRoutes registers protected balance endpoints.
func RegisterProtectedRoutes(
	protected *gin.RouterGroup,
	useCases modulefactory.UseCaseFactory,
	log port.Logger,
) {
	balanceHandler := balancehandler.NewBalanceHandler(useCases, log)
	protected.GET("/balance", balanceHandler.Get)
	protected.POST("/balance/withdraw", balanceHandler.Withdraw)
	protected.GET("/withdrawals", balanceHandler.ListWithdrawals)
}
