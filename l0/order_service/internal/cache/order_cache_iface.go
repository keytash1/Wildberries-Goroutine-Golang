package cache

import (
	"context"

	"order-service/internal/domain"
)

type OrderCache interface {
	Set(ctx context.Context, id string, order *domain.Order) error
	Get(ctx context.Context, id string) (*domain.Order, error)
	Restore(ctx context.Context, orders []*domain.Order) error
}
