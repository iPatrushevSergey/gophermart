package bootstrap

import (
	"context"
	"net/http"
	"time"

	adapterclock "gophermart/internal/gophermart/adapters/clock"
	"gophermart/internal/gophermart/adapters/repository/postgres"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/config"
	balancerepopostgres "gophermart/internal/gophermart/modules/balance/adapters/repository/postgres"
	balanceport "gophermart/internal/gophermart/modules/balance/application/port"
	balanceservice "gophermart/internal/gophermart/modules/balance/domain/service"
	balanceworker "gophermart/internal/gophermart/modules/balance/presentation/worker"
	identityauth "gophermart/internal/gophermart/modules/identity/adapters/auth"
	identityrepopostgres "gophermart/internal/gophermart/modules/identity/adapters/repository/postgres"
	identityport "gophermart/internal/gophermart/modules/identity/application/port"
	identityworker "gophermart/internal/gophermart/modules/identity/presentation/worker"
	ordersaccrual "gophermart/internal/gophermart/modules/orders/adapters/accrual"
	ordersrepopostgres "gophermart/internal/gophermart/modules/orders/adapters/repository/postgres"
	ordersvalidation "gophermart/internal/gophermart/modules/orders/adapters/validation"
	ordersport "gophermart/internal/gophermart/modules/orders/application/port"
	ordersworker "gophermart/internal/gophermart/modules/orders/presentation/worker"
)

// App holds the HTTP server and dependencies.
type App struct {
	Server  *http.Server
	workers []backgroundWorker
}

type repositories struct {
	userRepo       identityport.UserRepository
	orderRepo      ordersport.OrderRepository
	balanceRepo    balanceport.BalanceAccountRepository
	withdrawalRepo balanceport.WithdrawalRepository
}

type backgroundWorker interface {
	Start(ctx context.Context)
}

// NewApp wires dependencies and returns the application (composition root).
func NewApp(cfg config.Config, log port.Logger, transactor *postgres.Transactor) *App {
	hasher := identityauth.NewBCryptHasher(cfg.Auth.BCryptCost)
	tokens := identityauth.NewJWTProvider(cfg.Auth.JWTSecret, cfg.Auth.JWTTTL)
	luhnValidator := ordersvalidation.NewLuhnValidator()

	accrualClient := ordersaccrual.NewClientFromConfig(cfg.Accrual.Client)
	repos := newRepositories(transactor)

	balanceSvc := balanceservice.BalanceService{}
	clk := adapterclock.Real{}

	ucFactory := NewUseCaseFactory(
		WithUserRepo(repos.userRepo),
		WithOrderRepo(repos.orderRepo),
		WithBalanceRepo(repos.balanceRepo),
		WithWithdrawalRepo(repos.withdrawalRepo),
		WithHasher(hasher),
		WithTransactor(transactor),
		WithValidator(luhnValidator),
		WithAccrualClient(accrualClient),
		WithClock(clk),
		WithBalanceSvc(balanceSvc),
		WithLogger(log),
		WithBatchSize(cfg.Accrual.BatchSize),
		WithMaxWorkers(cfg.Accrual.MaxWorkers),
		WithOptimisticRetries(cfg.OptimisticRetries),
	)

	router := NewRouter(ucFactory, tokens, log)
	srv := newServer(cfg.Server.Address, router)
	workers := newBackgroundWorkers(ucFactory, log, cfg.Accrual.PollInterval)

	return &App{Server: srv, workers: workers}
}

func newRepositories(transactor *postgres.Transactor) repositories {
	return repositories{
		userRepo:       identityrepopostgres.NewUserRepository(transactor),
		orderRepo:      ordersrepopostgres.NewOrderRepository(transactor),
		balanceRepo:    balancerepopostgres.NewBalanceAccountRepository(transactor),
		withdrawalRepo: balancerepopostgres.NewWithdrawalRepository(transactor),
	}
}

func newServer(address string, router http.Handler) *http.Server {
	return &http.Server{
		Addr:    address,
		Handler: router,
	}
}

func newBackgroundWorkers(
	ucFactory UseCaseFactory,
	log port.Logger,
	pollInterval time.Duration,
) []backgroundWorker {
	identityWorkers := identityworker.BuildWorkers(identityworker.RegistryParams{})
	ordersWorkers := ordersworker.BuildWorkers(ordersworker.RegistryParams{
		UseCases:     ucFactory,
		Log:          log,
		PollInterval: pollInterval,
	})
	balanceWorkers := balanceworker.BuildWorkers(balanceworker.RegistryParams{})

	workers := make(
		[]backgroundWorker,
		0,
		len(identityWorkers)+len(ordersWorkers)+len(balanceWorkers),
	)
	for _, w := range identityWorkers {
		workers = append(workers, w)
	}
	for _, w := range ordersWorkers {
		workers = append(workers, w)
	}
	for _, w := range balanceWorkers {
		workers = append(workers, w)
	}
	return workers
}

// StartBackground starts all background workers.
func (a *App) StartBackground(ctx context.Context) {
	for _, w := range a.workers {
		w.Start(ctx)
	}
}
