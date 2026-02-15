package bootstrap

import (
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/application/usecase"
	"gophermart/internal/gophermart/domain/service"
	"gophermart/internal/gophermart/domain/vo"
	"gophermart/internal/gophermart/presentation/factory"
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

// NewUseCaseFactory builds the use case factory with all dependencies.
func NewUseCaseFactory(
	userRepo port.UserRepository,
	orderRepo port.OrderRepository,
	balanceRepo port.BalanceAccountRepository,
	withdrawalRepo port.WithdrawalRepository,
	hasher port.PasswordHasher,
	tokens port.TokenProvider,
	transactor port.Transactor,
	validator vo.OrderNumberValidator,
	accrualClient port.AccrualClient,
	userSvc service.UserService,
	balanceSvc service.BalanceService,
	orderSvc service.OrderService,
	withdrawalSvc service.WithdrawalService,
	log port.Logger,
	batchSize int,
	optimisticRetries int,
) factory.UseCaseFactory {
	return &useCaseFactory{
		register:        usecase.NewRegisterUser(userRepo, balanceRepo, transactor, hasher, userSvc, balanceSvc),
		login:           usecase.NewLoginUser(userRepo, hasher),
		uploadOrder:     usecase.NewUploadOrder(orderRepo, validator, orderSvc),
		listOrders:      usecase.NewListOrders(orderRepo),
		getBalance:      usecase.NewGetBalance(balanceRepo),
		withdraw:        usecase.NewWithdraw(balanceRepo, withdrawalRepo, transactor, validator, withdrawalSvc, optimisticRetries),
		listWithdrawals: usecase.NewListWithdrawals(withdrawalRepo),
		processAccrual:  usecase.NewProcessAccrual(orderRepo, balanceRepo, accrualClient, transactor, log, batchSize, optimisticRetries),
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
