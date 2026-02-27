package converter

import (
	"gophermart/internal/gophermart/adapters/repository/postgres/model"
	"gophermart/internal/gophermart/domain/entity"
)

//go:generate goverter gen .

// goverter:converter
// goverter:output:file withdrawal_gen.go
// goverter:output:package converter
// goverter:extend gophermart/internal/gophermart/adapters/repository/postgres/converter/convext:CopyTime
type WithdrawalConverter interface {
	ToEntity(source model.Withdrawal) entity.Withdrawal
	ToModel(source entity.Withdrawal) model.Withdrawal
}
