package bootstrap

import (
	"github.com/gin-gonic/gin"

	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/presentation/http/handler"
	"gophermart/internal/gophermart/presentation/http/middleware"
)

// NewRouter builds the Gin engine with all routes and middleware (composition root).
// Auth middleware applies only to routes registered inside the protected group.
func NewRouter(
	userHandler *handler.UserHandler,
	orderHandler *handler.OrderHandler,
	balanceHandler *handler.BalanceHandler,
	tokens port.TokenProvider,
	log port.Logger,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// Middleware with injected strategies (OCP compliant)
	r.Use(middleware.Compress(log, middleware.NewGzipCompressor()))
	r.Use(middleware.Logger(log, nil)) // nil uses DefaultLogFormatter

	api := r.Group("/api/user")
	{
		api.POST("/register", userHandler.Register)
		api.POST("/login", userHandler.Login)

		protected := api.Group("")
		protected.Use(middleware.Auth(nil, tokens)) // nil uses BearerTokenExtractor
		{
			protected.POST("/orders", orderHandler.Upload)
			protected.GET("/orders", orderHandler.List)
			protected.GET("/balance", balanceHandler.Get)
			protected.POST("/balance/withdraw", balanceHandler.Withdraw)
			protected.GET("/withdrawals", balanceHandler.ListWithdrawals)
		}
	}
	return r
}
