package handler

import (
	"net/http"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/presentation/factory"
	"gophermart/internal/gophermart/presentation/http/middleware"
	httpdto "gophermart/internal/gophermart/presentation/http/dto"

	"github.com/gin-gonic/gin"
)

// UserHandler handles POST /api/user/register and POST /api/user/login.
type UserHandler struct {
	factory factory.UseCaseFactory
	tokens  port.TokenProvider
}

// NewUserHandler creates a UserHandler with use case factory and token provider.
func NewUserHandler(factory factory.UseCaseFactory, tokens port.TokenProvider) *UserHandler {
	return &UserHandler{factory: factory, tokens: tokens}
}

// Register creates a new user and issues auth token.
func (h *UserHandler) Register(c *gin.Context) {
	var req httpdto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID, err := h.factory.RegisterUseCase().Execute(
		c.Request.Context(),
		dto.RegisterInput{Login: req.Login, Password: req.Password},
	)
	if err != nil {
		if err == application.ErrAlreadyExists {
			c.AbortWithStatus(http.StatusConflict)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	token, err := h.tokens.Issue(userID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	setAuthToken(c, token)
	c.Status(http.StatusOK)
}

// Login authenticates user and issues auth token.
func (h *UserHandler) Login(c *gin.Context) {
	var req httpdto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID, err := h.factory.LoginUseCase().Execute(
		c.Request.Context(),
		dto.LoginInput{Login: req.Login, Password: req.Password},
	)
	if err != nil {
		if err == application.ErrInvalidCredentials {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	token, err := h.tokens.Issue(userID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	setAuthToken(c, token)
	c.Status(http.StatusOK)
}

// setAuthToken writes the token to cookie and Authorization header.
func setAuthToken(c *gin.Context, token string) {
	cookie := &http.Cookie{
		Name:     middleware.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // true in prod (HTTPS)
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Writer, cookie)
	c.Header("Authorization", "Bearer "+token)
}
