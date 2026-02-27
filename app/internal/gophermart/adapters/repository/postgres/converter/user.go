package converter

import (
	"gophermart/internal/gophermart/adapters/repository/postgres/model"
	"gophermart/internal/gophermart/domain/entity"
)

//go:generate goverter gen .

// goverter:converter
// goverter:output:file user_gen.go
// goverter:output:package converter
// goverter:extend gophermart/internal/gophermart/adapters/repository/postgres/converter/convext:CopyTime
type UserConverter interface {
	ToEntity(source model.User) entity.User
	ToModel(source entity.User) model.User
}
