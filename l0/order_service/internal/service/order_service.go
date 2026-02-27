package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"order-service/internal/cache"
	"order-service/internal/domain"
	"order-service/internal/repository"
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrInvalidOrder  = errors.New("invalid order data")
)

var tracer = otel.Tracer("order-service")

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

func (s *OrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	ctx, span := tracer.Start(ctx, "OrderService.GetOrder")
	defer span.End()
	span.SetAttributes(attribute.String("order.id", id))

	//try cache
	order, err := s.cache.Get(ctx, id)
	if err != nil {
		log.Printf("Cache error for order %s: %v", id, err)
	} else if order != nil {
		//cache hit
		return order, nil
	}

	//try db
	order, err = s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}

	//save into cache
	if err := s.cache.Set(ctx, id, order); err != nil {
		log.Printf("failed to cache order %s after db read: %v", id, err)
	} else {
		log.Printf("Order %s cached successfully", order.OrderUID)
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

func (s *OrderService) saveToCacheWithRetry(ctx context.Context, id string, order *domain.Order) {
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		if err := s.cache.Set(ctx, id, order); err == nil {
			log.Printf("Order %s cached successfully after %d attempts", id, i+1)
			return
		} else if i < maxRetries-1 {
			log.Printf("Failed to cache order %s (attempt %d/%d): %v. Retrying...",
				id, i+1, maxRetries, err)
			time.Sleep(time.Millisecond * 100 * time.Duration(i+1))
		} else {
			log.Printf("Failed to cache order %s after %d attempts: %v",
				id, maxRetries, err)
		}
	}
}

func (s *OrderService) SaveOrder(ctx context.Context, order *domain.Order) error {
	ctx, span := tracer.Start(ctx, "OrderService.SaveOrder")
	defer span.End()

	span.SetAttributes(
		attribute.String("order.id", order.OrderUID),
		attribute.Int("items.count", len(order.Items)),
	)

	if err := s.validateOrder(order); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.type", "validation"))
		log.Printf("Invalid order %s: %v", order.OrderUID, err)
		return err
	}

	//save db
	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save to db: %w", err)
	}

	log.Printf("Order %s saved to database", order.OrderUID)

	//save cache
	go s.saveToCacheWithRetry(ctx, order.OrderUID, order)

	return nil
}

func (s *OrderService) RestoreCache(ctx context.Context) error {
	orders, err := s.repo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all orders: %w", err)
	}

	if err := s.cache.Restore(ctx, orders); err != nil {
		//лучше упасть?
		return fmt.Errorf("failed to restore cache: %w", err)
	}

	return nil
}
