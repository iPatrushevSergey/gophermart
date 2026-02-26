package middleware

import (
	"net/http"
	"strings"

	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/presentation/http/httpcontext"

	"github.com/gin-gonic/gin"
)

// TokenExtractor abstraction for retrieving auth token from request.
type TokenExtractor interface {
	Extract(c *gin.Context) (string, error)
}

// BearerTokenExtractor extracts token from "token" Cookie or "Authorization: Bearer" header.
type BearerTokenExtractor struct{}

func (e *BearerTokenExtractor) Extract(c *gin.Context) (string, error) {
	token := ""
	if t, err := c.Cookie(httpcontext.CookieName); err == nil && t != "" {
		token = t
	}
	if token == "" {
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}
	}
	return token, nil
}

// Auth middleware with injected strategy.
func Auth(extractor TokenExtractor, tokens port.TokenProvider) gin.HandlerFunc {
	if extractor == nil {
		extractor = &BearerTokenExtractor{}
	}
	return func(c *gin.Context) {
		token, err := extractor.Extract(c)
		if err != nil || token == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userID, err := tokens.Validate(token)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set(httpcontext.UserIDKey, userID)
		c.Next()
	}
}
