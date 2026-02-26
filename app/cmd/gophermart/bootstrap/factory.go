package bootstrap

import (
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/application/usecase"
	"gophermart/internal/gophermart/domain/service"
	"gophermart/internal/gophermart/domain/vo"
	"gophermart/internal/gophermart/presentation/factory"
	"gophermart/internal/pkg/option"
)

// useCaseFactory implements factory.UseCaseFactory; built in composition root.
type useCaseFactory struct {
	register        port.UseCase[dto.RegisterInput, vo.UserID]
	login           port.UseCase[dto.LoginInput, vo.UserID]
	uploadOrder     port.UseCase[dto.UploadOrderInput, struct{}]
	listOrders      port.UseCase[vo.UserID, []dto.OrderOutput]
	getBalance      port.UseCase[vo.UserID, dto.BalanceOutput]
	withdraw        port.UseCase[dto.WithdrawInput, struct{}]
	listWithdrawals port.UseCase[vo.UserID, []dto.WithdrawalOutput]
	processAccrual  port.BackgroundRunner
}

// factoryParams holds all dependencies needed to build the use case factory.
type factoryParams struct {
	userRepo          port.UserRepository
	orderRepo         port.OrderRepository
	balanceRepo       port.BalanceAccountRepository
	withdrawalRepo    port.WithdrawalRepository
	hasher            port.PasswordHasher
	tokens            port.TokenProvider
	transactor        port.Transactor
	validator         vo.OrderNumberValidator
	accrualClient     port.AccrualClient
	clock             port.Clock
	balanceSvc        service.BalanceService
	log               port.Logger
	batchSize         int
	maxWorkers        int
	optimisticRetries int
}

func WithUserRepo(r port.UserRepository) option.Option[factoryParams] {
	return func(p *factoryParams) { p.userRepo = r }
}

func WithOrderRepo(r port.OrderRepository) option.Option[factoryParams] {
	return func(p *factoryParams) { p.orderRepo = r }
}

func WithBalanceRepo(r port.BalanceAccountRepository) option.Option[factoryParams] {
	return func(p *factoryParams) { p.balanceRepo = r }
}

func WithWithdrawalRepo(r port.WithdrawalRepository) option.Option[factoryParams] {
	return func(p *factoryParams) { p.withdrawalRepo = r }
}

func WithHasher(h port.PasswordHasher) option.Option[factoryParams] {
	return func(p *factoryParams) { p.hasher = h }
}

func WithTokens(t port.TokenProvider) option.Option[factoryParams] {
	return func(p *factoryParams) { p.tokens = t }
}

func WithTransactor(t port.Transactor) option.Option[factoryParams] {
	return func(p *factoryParams) { p.transactor = t }
}

func WithValidator(v vo.OrderNumberValidator) option.Option[factoryParams] {
	return func(p *factoryParams) { p.validator = v }
}

func WithAccrualClient(c port.AccrualClient) option.Option[factoryParams] {
	return func(p *factoryParams) { p.accrualClient = c }
}

func WithClock(c port.Clock) option.Option[factoryParams] {
	return func(p *factoryParams) { p.clock = c }
}

func WithBalanceSvc(s service.BalanceService) option.Option[factoryParams] {
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
func NewUseCaseFactory(opts ...option.Option[factoryParams]) factory.UseCaseFactory {
	p := factoryParams{
		batchSize:         50,
		maxWorkers:        5,
		optimisticRetries: 3,
	}
	option.Apply(&p, opts...)

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
	if p.tokens == nil {
		panic("NewUseCaseFactory: WithTokens is required")
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

	return &useCaseFactory{
		register:        usecase.NewRegisterUser(p.userRepo, p.userRepo, p.balanceRepo, p.transactor, p.hasher, p.clock, p.balanceSvc),
		login:           usecase.NewLoginUser(p.userRepo, p.hasher),
		uploadOrder:     usecase.NewUploadOrder(p.orderRepo, p.orderRepo, p.validator, p.clock),
		listOrders:      usecase.NewListOrders(p.orderRepo),
		getBalance:      usecase.NewGetBalance(p.balanceRepo),
		withdraw:        usecase.NewWithdraw(p.balanceRepo, p.balanceRepo, p.withdrawalRepo, p.transactor, p.validator, p.clock, p.optimisticRetries),
		listWithdrawals: usecase.NewListWithdrawals(p.withdrawalRepo),
		processAccrual:  usecase.NewProcessAccrual(p.orderRepo, p.orderRepo, p.balanceRepo, p.balanceRepo, p.accrualClient, p.transactor, p.clock, p.log, p.batchSize, p.maxWorkers, p.optimisticRetries),
	}
}

func (f *useCaseFactory) RegisterUseCase() port.UseCase[dto.RegisterInput, vo.UserID] {
	return f.register
}

func (f *useCaseFactory) LoginUseCase() port.UseCase[dto.LoginInput, vo.UserID] {
	return f.login
}

func (f *useCaseFactory) UploadOrderUseCase() port.UseCase[dto.UploadOrderInput, struct{}] {
	return f.uploadOrder
}

func (f *useCaseFactory) ListOrdersUseCase() port.UseCase[vo.UserID, []dto.OrderOutput] {
	return f.listOrders
}

func (f *useCaseFactory) GetBalanceUseCase() port.UseCase[vo.UserID, dto.BalanceOutput] {
	return f.getBalance
}

func (f *useCaseFactory) WithdrawUseCase() port.UseCase[dto.WithdrawInput, struct{}] {
	return f.withdraw
}

func (f *useCaseFactory) ListWithdrawalsUseCase() port.UseCase[vo.UserID, []dto.WithdrawalOutput] {
	return f.listWithdrawals
}

func (f *useCaseFactory) ProcessAccrualUseCase() port.BackgroundRunner {
	return f.processAccrual
}
