package bootstrap

import (
	"github.com/gin-gonic/gin"

	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/presentation/http/handler"
	"gophermart/internal/gophermart/presentation/http/middleware"
)

// NewRouter builds the Gin engine with all routes and middleware (composition root).
// Auth middleware applies only to routes registered inside the protected group.
func NewRouter(userHandler *handler.UserHandler, tokens port.TokenProvider, log port.Logger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(log))
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
