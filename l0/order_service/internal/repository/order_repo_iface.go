package repository

import (
	"context"
	"order-service/internal/domain"
)

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	GetAll(ctx context.Context) ([]*domain.Order, error)
}
