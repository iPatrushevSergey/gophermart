package bootstrap

import (
	"net/http"

	"gophermart/internal/gophermart/adapters/accrual"
	"gophermart/internal/gophermart/adapters/auth"
	adapterclock "gophermart/internal/gophermart/adapters/clock"
	"gophermart/internal/gophermart/adapters/repository/postgres"
	"gophermart/internal/gophermart/adapters/validation"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/config"
	"gophermart/internal/gophermart/domain/service"
	"gophermart/internal/gophermart/presentation/http/handler"
	"gophermart/internal/gophermart/presentation/worker"
)

// App holds the HTTP server and dependencies.
type App struct {
	Server        *http.Server
	AccrualWorker *worker.AccrualWorker
}

// NewApp wires dependencies and returns the application (composition root).
func NewApp(cfg config.Config, log port.Logger, transactor *postgres.Transactor) *App {
	hasher := auth.NewBCryptHasher(cfg.Auth.BCryptCost)
	tokens := auth.NewJWTProvider(cfg.Auth.JWTSecret, cfg.Auth.JWTTTL)
	luhnValidator := validation.NewLuhnValidator()

	accrualHTTP := &http.Client{Timeout: cfg.Accrual.HTTPTimeout}
	accrualClient := accrual.NewClient(cfg.Accrual.Address, accrualHTTP)

	userRepo := postgres.NewUserRepository(transactor)
	orderRepo := postgres.NewOrderRepository(transactor)
	balanceRepo := postgres.NewBalanceAccountRepository(transactor)
	withdrawalRepo := postgres.NewWithdrawalRepository(transactor)

	// Domain services
	userSvc := service.UserService{}
	balanceSvc := service.BalanceService{}
	orderSvc := service.OrderService{}
	withdrawalSvc := service.WithdrawalService{}

	clk := adapterclock.Real{}

	ucFactory := NewUseCaseFactory(
		userRepo, orderRepo, balanceRepo, withdrawalRepo,
		hasher, tokens, transactor, luhnValidator, accrualClient, clk,
		userSvc, balanceSvc, orderSvc, withdrawalSvc, log,
		cfg.Accrual.BatchSize, cfg.Retry.OptimisticRetries,
	)

	userHandler := handler.NewUserHandler(ucFactory, tokens, log)
	orderHandler := handler.NewOrderHandler(ucFactory, log)
	balanceHandler := handler.NewBalanceHandler(ucFactory, log)

	router := NewRouter(userHandler, orderHandler, balanceHandler, tokens, log)

	srv := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}

	accrualWorker := worker.NewAccrualWorker(
		ucFactory.ProcessAccrualUseCase(), log, cfg.Accrual.PollInterval,
	)

	return &App{Server: srv, AccrualWorker: accrualWorker}
}
