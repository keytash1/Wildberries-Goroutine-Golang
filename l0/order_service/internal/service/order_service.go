package service

import (
	"errors"
	"fmt"
	"log"
	"order-service/internal/cache"
	"order-service/internal/domain"
	"order-service/internal/repository"
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrInvalidOrder  = errors.New("invalid order data")
)

type OrderService struct {
	repo  repository.OrderRepository
	cache cache.OrderCache
}

func NewOrderService(repo repository.OrderRepository, cache cache.OrderCache) *OrderService {
	return &OrderService{
		repo:  repo,
		cache: cache,
	}
}

func (s *OrderService) GetOrder(id string) (*domain.Order, error) {
	//try cache
	order, err := s.cache.Get(id)
	if err != nil {
		log.Printf("Cache error for order %s: %v", id, err)
	} else if order != nil {
		//cache hit
		return order, nil
	}

	//try db
	order, err = s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	//save into cache
	if err := s.cache.Set(id, order); err != nil {
		//log
		fmt.Printf("failed to set cache: %v\n", err)
	}

	return order, nil
}

func (s *OrderService) SaveOrder(order *domain.Order) error {
	//valid
	if order == nil {
		return ErrInvalidOrder
	}
	if order.OrderUID == "" {
		return fmt.Errorf("%w: missing order_uid", ErrInvalidOrder)
	}
	//save db
	if err := s.repo.Save(order); err != nil {
		return fmt.Errorf("failed to save to db: %w", err)
	}
	//save cache
	if err := s.cache.Set(order.OrderUID, order); err != nil {
		//log
		fmt.Printf("failed to set cache: %v\n", err)
	}
	log.Printf("Order recieved")
	return nil
}

func (s *OrderService) RestoreCache() error {
	orders, err := s.repo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get all orders: %w", err)
	}

	if err := s.cache.Restore(orders); err != nil {
		//лучше упасть?
		return fmt.Errorf("failed to restore cache: %w", err)
	}

	return nil
}
