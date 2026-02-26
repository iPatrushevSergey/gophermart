package httpcontext

import (
	"github.com/gin-gonic/gin"

	"gophermart/internal/gophermart/domain/vo"
)

// UserIDKey is the Gin context key for the authenticated user's ID.
const UserIDKey = "user_id"

// CookieName is the name of the auth cookie.
const CookieName = "token"

// UserID returns the authenticated user's ID from Gin context.
func UserID(c *gin.Context) (vo.UserID, bool) {
	v, ok := c.Get(UserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(vo.UserID)
	return id, ok
}
