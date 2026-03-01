package bootstrap

import (
	"gophermart/internal/gophermart/application/port"
	balanceapi "gophermart/internal/gophermart/modules/balance/application/api"
	balancedto "gophermart/internal/gophermart/modules/balance/application/dto"
	balancefactory "gophermart/internal/gophermart/modules/balance/application/factory"
	balanceport "gophermart/internal/gophermart/modules/balance/application/port"
	balanceservice "gophermart/internal/gophermart/modules/balance/domain/service"
	balancevo "gophermart/internal/gophermart/modules/balance/domain/vo"
	balancepresentationfactory "gophermart/internal/gophermart/modules/balance/presentation/factory"
	identityintermodule "gophermart/internal/gophermart/modules/identity/adapters/intermodule"
	identitydto "gophermart/internal/gophermart/modules/identity/application/dto"
	identityfactory "gophermart/internal/gophermart/modules/identity/application/factory"
	identityport "gophermart/internal/gophermart/modules/identity/application/port"
	identityvo "gophermart/internal/gophermart/modules/identity/domain/vo"
	identitypresentationfactory "gophermart/internal/gophermart/modules/identity/presentation/factory"
	ordersintermodule "gophermart/internal/gophermart/modules/orders/adapters/intermodule"
	ordersdto "gophermart/internal/gophermart/modules/orders/application/dto"
	ordersfactory "gophermart/internal/gophermart/modules/orders/application/factory"
	ordersport "gophermart/internal/gophermart/modules/orders/application/port"
	ordersvo "gophermart/internal/gophermart/modules/orders/domain/vo"
	orderspresentationfactory "gophermart/internal/gophermart/modules/orders/presentation/factory"
	"gophermart/internal/pkg/option"
)

// UseCaseFactory provides all module use cases needed by composition root.
type UseCaseFactory interface {
	identitypresentationfactory.UseCaseFactory
	orderspresentationfactory.UseCaseFactory
	balancepresentationfactory.UseCaseFactory
}

// useCaseFactory implements UseCaseFactory; built in composition root.
type useCaseFactory struct {
	register        port.UseCase[identitydto.RegisterInput, identityvo.UserID]
	login           port.UseCase[identitydto.LoginInput, identityvo.UserID]
	uploadOrder     port.UseCase[ordersdto.UploadOrderInput, struct{}]
	listOrders      port.UseCase[ordersvo.UserID, []ordersdto.OrderOutput]
	getBalance      port.UseCase[balancevo.UserID, balancedto.BalanceOutput]
	withdraw        port.UseCase[balancedto.WithdrawInput, struct{}]
	listWithdrawals port.UseCase[balancevo.UserID, []balancedto.WithdrawalOutput]
	processAccrual  port.BackgroundRunner
}

// factoryParams holds all dependencies needed to build the use case factory.
type factoryParams struct {
	userRepo          identityport.UserRepository
	orderRepo         ordersport.OrderRepository
	balanceRepo       balanceport.BalanceAccountRepository
	withdrawalRepo    balanceport.WithdrawalRepository
	hasher            port.PasswordHasher
	transactor        port.Transactor
	validator         ordersvo.OrderNumberValidator
	accrualClient     ordersport.AccrualClient
	clock             port.Clock
	balanceSvc        balanceservice.BalanceService
	log               port.Logger
	batchSize         int
	maxWorkers        int
	optimisticRetries int
}

func (p factoryParams) validate() {
	if p.userRepo == nil {
		panic("NewUseCaseFactory: WithUserRepo is required")
	}
	if p.orderRepo == nil {
		panic("NewUseCaseFactory: WithOrderRepo is required")
	}
	if p.balanceRepo == nil {
		panic("NewUseCaseFactory: WithBalanceRepo is required")
	}
	if p.withdrawalRepo == nil {
		panic("NewUseCaseFactory: WithWithdrawalRepo is required")
	}
	if p.hasher == nil {
		panic("NewUseCaseFactory: WithHasher is required")
	}
	if p.transactor == nil {
		panic("NewUseCaseFactory: WithTransactor is required")
	}
	if p.validator == nil {
		panic("NewUseCaseFactory: WithValidator is required")
	}
	if p.accrualClient == nil {
		panic("NewUseCaseFactory: WithAccrualClient is required")
	}
	if p.log == nil {
		panic("NewUseCaseFactory: WithLogger is required")
	}
}

func WithUserRepo(r identityport.UserRepository) option.Option[factoryParams] {
	return func(p *factoryParams) { p.userRepo = r }
}

func WithOrderRepo(r ordersport.OrderRepository) option.Option[factoryParams] {
	return func(p *factoryParams) { p.orderRepo = r }
}

func WithBalanceRepo(r balanceport.BalanceAccountRepository) option.Option[factoryParams] {
	return func(p *factoryParams) { p.balanceRepo = r }
}

func WithWithdrawalRepo(r balanceport.WithdrawalRepository) option.Option[factoryParams] {
	return func(p *factoryParams) { p.withdrawalRepo = r }
}

func WithHasher(h port.PasswordHasher) option.Option[factoryParams] {
	return func(p *factoryParams) { p.hasher = h }
}

func WithTransactor(t port.Transactor) option.Option[factoryParams] {
	return func(p *factoryParams) { p.transactor = t }
}

func WithValidator(v ordersvo.OrderNumberValidator) option.Option[factoryParams] {
	return func(p *factoryParams) { p.validator = v }
}

func WithAccrualClient(c ordersport.AccrualClient) option.Option[factoryParams] {
	return func(p *factoryParams) { p.accrualClient = c }
}

func WithClock(c port.Clock) option.Option[factoryParams] {
	return func(p *factoryParams) { p.clock = c }
}

func WithBalanceSvc(s balanceservice.BalanceService) option.Option[factoryParams] {
	return func(p *factoryParams) { p.balanceSvc = s }
}

func WithLogger(l port.Logger) option.Option[factoryParams] {
	return func(p *factoryParams) { p.log = l }
}

func WithBatchSize(n int) option.Option[factoryParams] {
	return func(p *factoryParams) { p.batchSize = n }
}

func WithMaxWorkers(n int) option.Option[factoryParams] {
	return func(p *factoryParams) { p.maxWorkers = n }
}

func WithOptimisticRetries(n int) option.Option[factoryParams] {
	return func(p *factoryParams) { p.optimisticRetries = n }
}

// NewUseCaseFactory builds the use case factory using functional options.
func NewUseCaseFactory(opts ...option.Option[factoryParams]) UseCaseFactory {
	p := factoryParams{
		batchSize:         50,
		maxWorkers:        5,
		optimisticRetries: 3,
	}
	option.Apply(&p, opts...)
	p.validate()

	balanceUC := buildBalanceUseCases(p)
	identityUC := buildIdentityUseCases(p, balanceUC.OpenAccount)
	ordersUC := buildOrdersUseCases(p, balanceUC.ApplyAccrual)

	return &useCaseFactory{
		register:        identityUC.Register,
		login:           identityUC.Login,
		uploadOrder:     ordersUC.UploadOrder,
		listOrders:      ordersUC.ListOrders,
		getBalance:      balanceUC.GetBalance,
		withdraw:        balanceUC.Withdraw,
		listWithdrawals: balanceUC.ListWithdrawals,
		processAccrual:  ordersUC.ProcessAccrual,
	}
}

func (f *useCaseFactory) RegisterUseCase() port.UseCase[identitydto.RegisterInput, identityvo.UserID] {
	return f.register
}

func (f *useCaseFactory) LoginUseCase() port.UseCase[identitydto.LoginInput, identityvo.UserID] {
	return f.login
}

func (f *useCaseFactory) UploadOrderUseCase() port.UseCase[ordersdto.UploadOrderInput, struct{}] {
	return f.uploadOrder
}

func (f *useCaseFactory) ListOrdersUseCase() port.UseCase[ordersvo.UserID, []ordersdto.OrderOutput] {
	return f.listOrders
}

func (f *useCaseFactory) GetBalanceUseCase() port.UseCase[balancevo.UserID, balancedto.BalanceOutput] {
	return f.getBalance
}

func (f *useCaseFactory) WithdrawUseCase() port.UseCase[balancedto.WithdrawInput, struct{}] {
	return f.withdraw
}

func (f *useCaseFactory) ListWithdrawalsUseCase() port.UseCase[balancevo.UserID, []balancedto.WithdrawalOutput] {
	return f.listWithdrawals
}

func (f *useCaseFactory) ProcessAccrualUseCase() port.BackgroundRunner {
	return f.processAccrual
}

func (p factoryParams) balanceParams() balancefactory.Params {
	return balancefactory.Params{
		BalanceRepo:       p.balanceRepo,
		WithdrawalRepo:    p.withdrawalRepo,
		Transactor:        p.transactor,
		Validator:         p.validator,
		Clock:             p.clock,
		BalanceSvc:        p.balanceSvc,
		OptimisticRetries: p.optimisticRetries,
	}
}

func buildBalanceUseCases(p factoryParams) balancefactory.UseCases {
	return balancefactory.NewUseCases(p.balanceParams())
}

func (p factoryParams) identityParams(balanceAccountAPI balanceapi.AccountAPI) identityfactory.Params {
	return identityfactory.Params{
		UserRepo:       p.userRepo,
		BalanceGateway: identityintermodule.NewBalanceGatewayAdapter(balanceAccountAPI),
		Transactor:     p.transactor,
		Hasher:         p.hasher,
		Clock:          p.clock,
	}
}

func buildIdentityUseCases(p factoryParams, balanceAccountAPI balanceapi.AccountAPI) identityfactory.UseCases {
	return identityfactory.NewUseCases(p.identityParams(balanceAccountAPI))
}

func (p factoryParams) ordersParams(balanceAccrualAPI balanceapi.AccrualAPI) ordersfactory.Params {
	return ordersfactory.Params{
		OrderRepo:         p.orderRepo,
		BalanceGateway:    ordersintermodule.NewBalanceGatewayAdapter(balanceAccrualAPI),
		Validator:         p.validator,
		AccrualClient:     p.accrualClient,
		Transactor:        p.transactor,
		Clock:             p.clock,
		Log:               p.log,
		BatchSize:         p.batchSize,
		MaxWorkers:        p.maxWorkers,
		OptimisticRetries: p.optimisticRetries,
	}
}

func buildOrdersUseCases(p factoryParams, balanceAccrualAPI balanceapi.AccrualAPI) ordersfactory.UseCases {
	return ordersfactory.NewUseCases(p.ordersParams(balanceAccrualAPI))
}
