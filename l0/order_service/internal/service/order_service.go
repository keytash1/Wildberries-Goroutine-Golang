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

func (s *OrderService) validateOrder(order *domain.Order) error {
	if order == nil {
		return fmt.Errorf("%w: order is nil", ErrInvalidOrder)
	}

	// orders table
	if order.OrderUID == "" {
		return fmt.Errorf("%w: missing order_uid", ErrInvalidOrder)
	}
	if order.TrackNumber == "" {
		return fmt.Errorf("%w: missing track_number", ErrInvalidOrder)
	}
	if order.CustomerID == "" {
		return fmt.Errorf("%w: missing customer_id", ErrInvalidOrder)
	}
	if order.DateCreated.IsZero() {
		return fmt.Errorf("%w: missing date_created", ErrInvalidOrder)
	}

	// delivery table
	if order.Delivery.Name == "" {
		return fmt.Errorf("%w: missing delivery name", ErrInvalidOrder)
	}
	if order.Delivery.Phone == "" {
		return fmt.Errorf("%w: missing delivery phone", ErrInvalidOrder)
	}
	if order.Delivery.City == "" {
		return fmt.Errorf("%w: missing delivery city", ErrInvalidOrder)
	}
	if order.Delivery.Address == "" {
		return fmt.Errorf("%w: missing delivery address", ErrInvalidOrder)
	}
	if order.Delivery.Email == "" {
		return fmt.Errorf("%w: missing delivery email", ErrInvalidOrder)
	}

	// payment table
	if order.Payment.Transaction == "" {
		return fmt.Errorf("%w: missing payment transaction", ErrInvalidOrder)
	}
	if order.Payment.Currency == "" {
		return fmt.Errorf("%w: missing payment currency", ErrInvalidOrder)
	}
	if order.Payment.Provider == "" {
		return fmt.Errorf("%w: missing payment provider", ErrInvalidOrder)
	}
	if order.Payment.Amount <= 0 {
		return fmt.Errorf("%w: invalid payment amount", ErrInvalidOrder)
	}

	// items table
	if len(order.Items) == 0 {
		return fmt.Errorf("%w: no items in order", ErrInvalidOrder)
	}
	for i, item := range order.Items {
		if item.ChrtID <= 0 {
			return fmt.Errorf("%w: item %d: invalid chrt_id", ErrInvalidOrder, i)
		}
		if item.Price <= 0 {
			return fmt.Errorf("%w: item %d: invalid price", ErrInvalidOrder, i)
		}
		if item.Rid == "" {
			return fmt.Errorf("%w: item %d: missing rid", ErrInvalidOrder, i)
		}
		if item.Name == "" {
			return fmt.Errorf("%w: item %d: missing name", ErrInvalidOrder, i)
		}
		if item.TotalPrice <= 0 {
			return fmt.Errorf("%w: item %d: invalid total_price", ErrInvalidOrder, i)
		}
	}

	return nil
}

func (s *OrderService) SaveOrder(order *domain.Order) error {
	if err := s.validateOrder(order); err != nil {
		log.Printf("Invalid order: %v", err)
		return err
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
