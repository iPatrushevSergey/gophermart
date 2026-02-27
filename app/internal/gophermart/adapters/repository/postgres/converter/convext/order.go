package convext

import (
	"fmt"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

func StatusFromDB(v int16) (entity.OrderStatus, error) {
	switch v {
	case 0:
		return entity.OrderStatusNew, nil
	case 1:
		return entity.OrderStatusProcessing, nil
	case 2:
		return entity.OrderStatusInvalid, nil
	case 3:
		return entity.OrderStatusProcessed, nil
	default:
		return "", fmt.Errorf("unknown order status code: %d", v)
	}
}

func StatusToDB(v entity.OrderStatus) (int16, error) {
	switch v {
	case entity.OrderStatusNew:
		return 0, nil
	case entity.OrderStatusProcessing:
		return 1, nil
	case entity.OrderStatusInvalid:
		return 2, nil
	case entity.OrderStatusProcessed:
		return 3, nil
	default:
		return 0, fmt.Errorf("unknown order status: %s", v)
	}
}

func PointsFromDB(v *float64) *vo.Points {
	if v == nil {
		return nil
	}
	p := vo.Points(*v)
	return &p
}

func PointsToDB(v *vo.Points) *float64 {
	if v == nil {
		return nil
	}
	f := float64(*v)
	return &f
}
