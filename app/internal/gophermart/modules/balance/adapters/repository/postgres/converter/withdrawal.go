package converter

import (
	"gophermart/internal/gophermart/modules/balance/adapters/repository/postgres/model"
	"gophermart/internal/gophermart/modules/balance/domain/entity"
)

//go:generate goverter gen .

// goverter:converter
// goverter:output:file withdrawal_gen.go
// goverter:output:package converter
// goverter:extend gophermart/internal/gophermart/modules/balance/adapters/repository/postgres/converter/convext:CopyTime
type WithdrawalConverter interface {
	ToEntity(source model.Withdrawal) entity.Withdrawal
	ToModel(source entity.Withdrawal) model.Withdrawal
}
