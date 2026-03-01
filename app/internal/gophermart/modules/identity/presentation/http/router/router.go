package router

import (
	"github.com/gin-gonic/gin"

	"gophermart/internal/gophermart/application/port"
	modulefactory "gophermart/internal/gophermart/modules/identity/presentation/factory"
	identityhandler "gophermart/internal/gophermart/modules/identity/presentation/http/handler"
)

// RegisterPublicRoutes registers public identity endpoints.
func RegisterPublicRoutes(
	api *gin.RouterGroup,
	useCases modulefactory.UseCaseFactory,
	tokens port.TokenProvider,
	log port.Logger,
) {
	userHandler := identityhandler.NewUserHandler(useCases, tokens, log)
	api.POST("/register", userHandler.Register)
	api.POST("/login", userHandler.Login)
}
