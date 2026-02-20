package repository

import (
	"context"
	"fmt"
	"log"
	"order-service/internal/config"
	"order-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOrderRepo struct {
	pool *pgxpool.Pool
}

// return interface
func NewPostgresOrderRepo(cfg config.DBConfig) (*PostgresOrderRepo, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	return &PostgresOrderRepo{pool: pool}, nil
}

func (r *PostgresOrderRepo) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

func (r *PostgresOrderRepo) Save(order *domain.Order) error {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	//сохранили заказ
	_, err = tx.Exec(ctx, `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES  ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.Shardkey, order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	//сохранили доставку
	_, err = tx.Exec(ctx, `
        INSERT INTO delivery (
            order_uid, name, phone, zip, city, address, region, email
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone,
		order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("failed to save delivery: %w", err)
	}

	//сохраняем платеж
	_, err = tx.Exec(ctx, `
        INSERT INTO payment (
            order_uid, transaction, request_id, currency, provider, amount,
            payment_dt, bank, delivery_cost, goods_total, custom_fee
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	//добавляем новые товары
	for i, item := range order.Items {
		_, err = tx.Exec(ctx, `
            INSERT INTO items (
                order_uid, chrt_id, track_number, price, rid, name,
                sale, size, total_price, nm_id, brand, status
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        `,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid,
			item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID,
			item.Brand, item.Status,
		)
		if err != nil {
			return fmt.Errorf("failed to save item %d:%w", i, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	log.Printf("Message saved")
	return nil
}

func (r *PostgresOrderRepo) GetByID(id string) (*domain.Order, error) {
	ctx := context.Background()
	order := &domain.Order{Items: []domain.Item{}}

	//получаем заказ
	err := r.pool.QueryRow(ctx, `
        SELECT order_uid, track_number, entry, locale, internal_signature,
               customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders WHERE order_uid = $1
    `, id).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	//получаем доставку
	err = r.pool.QueryRow(ctx, `
        SELECT name, phone, zip, city, address, region, email
        FROM delivery WHERE order_uid = $1
    `, id).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
		&order.Delivery.Email,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}

	//получаем платеж
	err = r.pool.QueryRow(ctx, `
        SELECT transaction, request_id, currency, provider, amount, payment_dt,
               bank, delivery_cost, goods_total, custom_fee
        FROM payment WHERE order_uid = $1
    `, id).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID,
		&order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount,
		&order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal, &order.Payment.CustomFee,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	//получаем товары
	rows, err := r.pool.Query(ctx, `
        SELECT chrt_id, track_number, price, rid, name, sale, size,
               total_price, nm_id, brand, status
        FROM items WHERE order_uid = $1
        ORDER BY id
    `, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.Item
		err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice,
			&item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

// getAll для восстановления кэша
// ПРОБЛЕМА N+1, ПЕРЕДЕЛАТЬ
func (r *PostgresOrderRepo) GetAll() ([]*domain.Order, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, "SELECT order_uid FROM orders ORDER BY date_created DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to get order list %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		//плохо
		order, err := r.GetByID(id)
		if err == nil && order != nil {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

func (r *PostgresOrderRepo) Ping() error {
	ctx := context.Background()
	return r.pool.Ping(ctx)
}
