package cache

import "order-service/internal/domain"

type OrderCache interface {
	Set(id string, order *domain.Order) error
	Get(id string) (*domain.Order, error)
	Restore(orders []*domain.Order) error
}
