package converter

import (
	"gophermart/internal/gophermart/modules/orders/adapters/repository/postgres/model"
	"gophermart/internal/gophermart/modules/orders/domain/entity"
)

//go:generate goverter gen .

// goverter:converter
// goverter:output:file order_gen.go
// goverter:output:package converter
// goverter:enum:unknown @error
// goverter:extend gophermart/internal/gophermart/modules/orders/adapters/repository/postgres/converter/convext:StatusFromDB
// goverter:extend gophermart/internal/gophermart/modules/orders/adapters/repository/postgres/converter/convext:StatusToDB
// goverter:extend gophermart/internal/gophermart/modules/orders/adapters/repository/postgres/converter/convext:PointsFromDB
// goverter:extend gophermart/internal/gophermart/modules/orders/adapters/repository/postgres/converter/convext:PointsToDB
// goverter:extend gophermart/internal/gophermart/modules/orders/adapters/repository/postgres/converter/convext:CopyTime
// goverter:extend gophermart/internal/gophermart/modules/orders/adapters/repository/postgres/converter/convext:CopyTimePtr
type OrderConverter interface {
	ToEntity(source model.Order) (entity.Order, error)
	ToModel(source entity.Order) (model.Order, error)
}
