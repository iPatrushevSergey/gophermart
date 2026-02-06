package bootstrap

import (
	"net/http"

	"gophermart/internal/gophermart/adapters/auth"
	"gophermart/internal/gophermart/adapters/repository/fake"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/config"
	"gophermart/internal/gophermart/domain/service"
	"gophermart/internal/gophermart/presentation/http/handler"
)

// App holds the HTTP server and dependencies.
type App struct {
	Server *http.Server
}

// NewApp wires dependencies and returns the application (composition root).
func NewApp(cfg config.Config, log port.Logger) *App {
	userRepo := fake.NewUserRepository()
	hasher := auth.NewBCryptHasher()
	tokens := auth.NewJWTProvider(cfg.JWTSecret, cfg.JWTTTL)
	userSvc := service.UserService{}
	ucFactory := NewUseCaseFactory(userRepo, hasher, tokens, userSvc)
	userHandler := handler.NewUserHandler(ucFactory, tokens, log)
	router := NewRouter(userHandler, tokens, log)

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}
	return &App{Server: srv}
}
