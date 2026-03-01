package bootstrap

import (
	"github.com/gin-gonic/gin"

	"gophermart/internal/gophermart/application/port"
	balancerouter "gophermart/internal/gophermart/modules/balance/presentation/http/router"
	identityport "gophermart/internal/gophermart/modules/identity/application/port"
	identityrouter "gophermart/internal/gophermart/modules/identity/presentation/http/router"
	ordersrouter "gophermart/internal/gophermart/modules/orders/presentation/http/router"
	"gophermart/internal/gophermart/presentation/http/middleware"
)

type identityTokenValidatorBridge struct {
	tokens identityport.TokenProvider
}

func (a identityTokenValidatorBridge) Validate(token string) (int64, error) {
	userID, err := a.tokens.Validate(token)
	if err != nil {
		return 0, err
	}
	return int64(userID), nil
}

// NewRouter builds the Gin engine with all routes and middleware (composition root).
// Auth middleware applies only to routes registered inside the protected group.
func NewRouter(
	useCases UseCaseFactory,
	tokens identityport.TokenProvider,
	log port.Logger,
) *gin.Engine {
	r := gin.New()
	globalParams := middleware.GlobalRegistryParams{
		Log:    log,
		Tokens: identityTokenValidatorBridge{tokens: tokens},
	}
	r.Use(middleware.BuildAppMiddleware(globalParams)...)

	api := r.Group("/api/user")
	{
		identityrouter.RegisterPublicRoutes(api, useCases, tokens, log)

		protected := api.Group("")
		protected.Use(middleware.BuildProtectedMiddleware(globalParams)...)
		{
			ordersrouter.RegisterProtectedRoutes(protected, useCases, log)
			balancerouter.RegisterProtectedRoutes(protected, useCases, log)
		}
	}
	return r
}
