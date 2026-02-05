package bootstrap

import (
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/presentation/http/handler"
	"gophermart/internal/gophermart/presentation/http/middleware"
	"github.com/gin-gonic/gin"
)

// NewRouter builds the Gin engine with all routes and middleware (composition root).
// Auth middleware applies only to routes registered inside the protected group.
func NewRouter(userHandler *handler.UserHandler, tokens port.TokenProvider) *gin.Engine {
	r := gin.Default()
	api := r.Group("/api/user")
	{
		api.POST("/register", userHandler.Register)
		api.POST("/login", userHandler.Login)
		protected := api.Group("")
		protected.Use(middleware.Auth(tokens))
		{
			// protected.GET("/orders", ...)
			// protected.GET("/balance", ...)
			// protected.POST("/orders", ...)
			// protected.POST("/balance/withdraw", ...)
			// protected.GET("/withdrawals", ...)
		}
	}
	return r
}
