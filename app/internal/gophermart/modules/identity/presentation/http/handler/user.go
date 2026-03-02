package handler

import (
	"net/http"

	"gophermart/internal/gophermart/application"
	appport "gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/modules/identity/application/dto"
	"gophermart/internal/gophermart/modules/identity/application/port"
	"gophermart/internal/gophermart/modules/identity/presentation/factory"
	httpdto "gophermart/internal/gophermart/modules/identity/presentation/http/dto"
	"gophermart/internal/gophermart/presentation/http/httpcontext"

	"github.com/gin-gonic/gin"
)

// UserHandler manages registration and authentication requests.
type UserHandler struct {
	useCases factory.UseCaseFactory
	tokens   port.TokenProvider
	log      appport.Logger
}

// NewUserHandler creates a UserHandler with identity use cases provider.
func NewUserHandler(useCases factory.UseCaseFactory, tokens port.TokenProvider, log appport.Logger) *UserHandler {
	return &UserHandler{
		useCases: useCases,
		tokens:   tokens,
		log:      log,
	}
}

// Register creates a new user and issues an auth token.
func (h *UserHandler) Register(c *gin.Context) {
	var req httpdto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID, err := h.useCases.RegisterUseCase().Execute(
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

// Login authenticates a user and issues an auth token.
func (h *UserHandler) Login(c *gin.Context) {
	var req httpdto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID, err := h.useCases.LoginUseCase().Execute(
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
		Name:     httpcontext.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // true in prod (HTTPS)
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Writer, cookie)
	c.Header("Authorization", "Bearer "+token)
}
