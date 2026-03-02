package router

import (
	"github.com/gin-gonic/gin"

	appport "gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/modules/identity/application/port"
	"gophermart/internal/gophermart/modules/identity/presentation/factory"
	"gophermart/internal/gophermart/modules/identity/presentation/http/handler"
)

// RegisterPublicRoutes registers public identity endpoints.
func RegisterPublicRoutes(
	api *gin.RouterGroup,
	useCases factory.UseCaseFactory,
	tokens port.TokenProvider,
	log appport.Logger,
) {
	userHandler := handler.NewUserHandler(useCases, tokens, log)
	api.POST("/register", userHandler.Register)
	api.POST("/login", userHandler.Login)
}
