package service

import "order-service/internal/domain"

type OrderServiceInterface interface {
	SaveOrder(order *domain.Order) error
}
