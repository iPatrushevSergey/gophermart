package handler

import (
	"net/http"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/presentation/factory"
	httpdto "gophermart/internal/gophermart/presentation/http/dto"
	"gophermart/internal/gophermart/presentation/http/middleware"

	"github.com/gin-gonic/gin"
)

// UserHandler handles POST /api/user/register and POST /api/user/login.
type UserHandler struct {
	factory factory.UseCaseFactory
	tokens  port.TokenProvider
	log     port.Logger
}

// NewUserHandler creates a UserHandler with use case factory, token provider and logger.
func NewUserHandler(factory factory.UseCaseFactory, tokens port.TokenProvider, log port.Logger) *UserHandler {
	return &UserHandler{factory: factory, tokens: tokens, log: log}
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
		h.log.Error("register use case failed", "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	token, err := h.tokens.Issue(userID)
	if err != nil {
		h.log.Error("failed to issue token", "error", err)
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
		h.log.Error("login use case failed", "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	token, err := h.tokens.Issue(userID)
	if err != nil {
		h.log.Error("failed to issue token", "error", err)
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
