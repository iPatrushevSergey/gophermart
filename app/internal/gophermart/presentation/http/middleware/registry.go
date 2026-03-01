package middleware

import (
	"github.com/gin-gonic/gin"

	"gophermart/internal/gophermart/application/port"
)

// GlobalRegistryParams contains dependencies required to build global middleware.
type GlobalRegistryParams struct {
	Log    port.Logger
	Tokens port.TokenProvider
}

// BuildAppMiddleware builds middleware for the whole HTTP app.
func BuildAppMiddleware(p GlobalRegistryParams) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		gin.Recovery(),
		Compress(p.Log, NewGzipCompressor()),
		Logger(p.Log, nil),
	}
}

// BuildProtectedMiddleware builds middleware for protected API routes.
func BuildProtectedMiddleware(p GlobalRegistryParams) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		Auth(nil, p.Tokens),
	}
}
