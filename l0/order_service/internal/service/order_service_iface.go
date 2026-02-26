package service

import (
	"context"

	"order-service/internal/domain"
)

type OrderServiceInterface interface {
	GetOrder(ctx context.Context, id string) (*domain.Order, error)
	SaveOrder(ctx context.Context, order *domain.Order) error
	RestoreCache(ctx context.Context) error
}
