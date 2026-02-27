package converter

import (
	"gophermart/internal/gophermart/adapters/repository/postgres/model"
	"gophermart/internal/gophermart/domain/entity"
)

//go:generate goverter gen .

// goverter:converter
// goverter:output:file balance_account_gen.go
// goverter:output:package converter
// goverter:extend gophermart/internal/gophermart/adapters/repository/postgres/converter/convext:CopyTime
type BalanceAccountConverter interface {
	ToEntity(source model.BalanceAccount) entity.BalanceAccount
	ToModel(source entity.BalanceAccount) model.BalanceAccount
}
