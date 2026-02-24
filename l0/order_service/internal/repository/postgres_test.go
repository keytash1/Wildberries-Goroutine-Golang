package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"order-service/internal/domain"
)

func TestPostgresOrderRepo_Save_SuccessAndIdempotent(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &PostgresOrderRepo{pool: mock}

	now := time.Now().UTC().Truncate(time.Second)

	order := &domain.Order{
		OrderUID:          "b563feb7b2b84b6test",
		TrackNumber:       "WBILMTESTTRACK",
		Entry:             "WBIL",
		Locale:            "ru",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       now,
		OofShard:          "1",
		Delivery: domain.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: domain.Payment{
			Transaction:  "b563feb7b2b84b6test",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []domain.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
	}

	// первый вызов Save — должен вставить все записи
	mock.ExpectBegin()

	// orders
	mock.ExpectExec(`INSERT INTO orders`).
		WithArgs(
			order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
			order.InternalSignature, order.CustomerID, order.DeliveryService,
			order.Shardkey, order.SmID, order.DateCreated, order.OofShard,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// delivery
	mock.ExpectExec(`INSERT INTO delivery`).
		WithArgs(
			order.OrderUID, order.Delivery.Name, order.Delivery.Phone,
			order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
			order.Delivery.Region, order.Delivery.Email,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// payment
	mock.ExpectExec(`INSERT INTO payment`).
		WithArgs(
			order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
			order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
			order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
			order.Payment.GoodsTotal, order.Payment.CustomFee,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// items
	mock.ExpectExec(`INSERT INTO items`).
		WithArgs(
			order.OrderUID,
			order.Items[0].ChrtID,
			order.Items[0].TrackNumber,
			order.Items[0].Price,
			order.Items[0].Rid,
			order.Items[0].Name,
			order.Items[0].Sale,
			order.Items[0].Size,
			order.Items[0].TotalPrice,
			order.Items[0].NmID,
			order.Items[0].Brand,
			order.Items[0].Status,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mock.ExpectCommit()

	err = repo.Save(order)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())

	// Второй вызов того же заказа ON CONFLICT DO NOTHING
	// Ожидаем те же запросы, но RowsAffected = 0
	mock.ExpectBegin()

	mock.ExpectExec(`INSERT INTO orders`).
		WithArgs(
			order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
			order.InternalSignature, order.CustomerID, order.DeliveryService,
			order.Shardkey, order.SmID, order.DateCreated, order.OofShard,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 0))

	mock.ExpectExec(`INSERT INTO delivery`).
		WithArgs(
			order.OrderUID, order.Delivery.Name, order.Delivery.Phone,
			order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
			order.Delivery.Region, order.Delivery.Email,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 0))

	mock.ExpectExec(`INSERT INTO payment`).
		WithArgs(
			order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
			order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
			order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
			order.Payment.GoodsTotal, order.Payment.CustomFee,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 0))

	mock.ExpectExec(`INSERT INTO items`).
		WithArgs(
			order.OrderUID,
			order.Items[0].ChrtID,
			order.Items[0].TrackNumber,
			order.Items[0].Price,
			order.Items[0].Rid,
			order.Items[0].Name,
			order.Items[0].Sale,
			order.Items[0].Size,
			order.Items[0].TotalPrice,
			order.Items[0].NmID,
			order.Items[0].Brand,
			order.Items[0].Status,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 0))

	mock.ExpectCommit()

	err = repo.Save(order)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresOrderRepo_Save_BeginError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &PostgresOrderRepo{pool: mock}

	mock.ExpectBegin().WillReturnError(errors.New("cannot begin transaction"))

	order := &domain.Order{OrderUID: "test-error"}

	err = repo.Save(order)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresOrderRepo_GetByID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &PostgresOrderRepo{pool: mock}

	now := time.Now().UTC().Truncate(time.Second)

	orderRows := mock.NewRows([]string{
		"order_uid", "track_number", "entry", "locale", "internal_signature",
		"customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard",
	}).AddRow(
		"b563feb7b2b84b6test", "WBILMTESTTRACK", "WBIL", "ru", "",
		"test", "meest", "9", int64(99), now, "1",
	)

	mock.ExpectQuery(`SELECT order_uid, track_number, entry, locale, internal_signature,
               customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders WHERE order_uid = \$1`).
		WithArgs("b563feb7b2b84b6test").
		WillReturnRows(orderRows)

	deliveryRows := mock.NewRows([]string{
		"name", "phone", "zip", "city", "address", "region", "email",
	}).AddRow(
		"Test Testov", "+9720000000", "2639809", "Kiryat Mozkin",
		"Ploshad Mira 15", "Kraiot", "test@gmail.com",
	)

	mock.ExpectQuery(`SELECT name, phone, zip, city, address, region, email
        FROM delivery WHERE order_uid = \$1`).
		WithArgs("b563feb7b2b84b6test").
		WillReturnRows(deliveryRows)

	paymentRows := mock.NewRows([]string{
		"transaction", "request_id", "currency", "provider", "amount",
		"payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee",
	}).AddRow(
		"b563feb7b2b84b6test", "", "USD", "wbpay", 1817,
		int64(1637907727), "alpha", 1500, 317, 0,
	)

	mock.ExpectQuery(`SELECT transaction, request_id, currency, provider, amount, payment_dt,
               bank, delivery_cost, goods_total, custom_fee
        FROM payment WHERE order_uid = \$1`).
		WithArgs("b563feb7b2b84b6test").
		WillReturnRows(paymentRows)

	itemsRows := mock.NewRows([]string{
		"chrt_id", "track_number", "price", "rid", "name", "sale",
		"size", "total_price", "nm_id", "brand", "status",
	}).AddRow(
		int64(9934930),          // chrt_id
		"WBILMTESTTRACK",        // track_number
		453,                     // price (int)
		"ab4219087a764ae0btest", // rid
		"Mascaras",              // name
		30,                      // sale (int)
		"0",                     // size
		317,                     // total_price (int)
		int64(2389212),          // nm_id
		"Vivienne Sabo",         // brand
		202,                     // status (int)
	)

	mock.ExpectQuery(`SELECT chrt_id, track_number, price, rid, name, sale, size,
               total_price, nm_id, brand, status
        FROM items WHERE order_uid = \$1
        ORDER BY id`).
		WithArgs("b563feb7b2b84b6test").
		WillReturnRows(itemsRows)

	order, err := repo.GetByID("b563feb7b2b84b6test")
	require.NoError(t, err)
	require.NotNil(t, order)

	assert.Equal(t, "b563feb7b2b84b6test", order.OrderUID)
	assert.Equal(t, "Test Testov", order.Delivery.Name)
	assert.Equal(t, 1817, order.Payment.Amount)
	assert.Len(t, order.Items, 1)
	assert.Equal(t, "Mascaras", order.Items[0].Name)
	assert.Equal(t, 453, order.Items[0].Price)
	assert.Equal(t, 30, order.Items[0].Sale)
	assert.Equal(t, 317, order.Items[0].TotalPrice)
	assert.Equal(t, 202, order.Items[0].Status)
	assert.Equal(t, int64(9934930), order.Items[0].ChrtID)
	assert.Equal(t, int64(2389212), order.Items[0].NmID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresOrderRepo_GetByID_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &PostgresOrderRepo{pool: mock}

	mock.ExpectQuery(`SELECT .+ FROM orders WHERE order_uid = \$1`).
		WithArgs("non-existent").
		WillReturnError(pgx.ErrNoRows)

	order, err := repo.GetByID("non-existent")
	require.NoError(t, err)
	assert.Nil(t, order)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresOrderRepo_GetByID_OrderError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &PostgresOrderRepo{pool: mock}

	mock.ExpectQuery(`SELECT .+ FROM orders WHERE order_uid = \$1`).
		WithArgs("test-error").
		WillReturnError(errors.New("connection error"))

	order, err := repo.GetByID("test-error")
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "failed to get order")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresOrderRepo_GetAll_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &PostgresOrderRepo{pool: mock}

	now := time.Now().UTC().Truncate(time.Second)

	idRows := mock.NewRows([]string{"order_uid"}).
		AddRow("order-1").
		AddRow("order-2")

	mock.ExpectQuery(`SELECT order_uid FROM orders ORDER BY date_created DESC`).
		WillReturnRows(idRows)

	order1Rows := mock.NewRows([]string{
		"order_uid", "track_number", "entry", "locale", "internal_signature",
		"customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard",
	}).AddRow("order-1", "TRACK1", "WBIL", "ru", "", "cust1", "meest", "9", int64(99), now, "1")

	mock.ExpectQuery(`SELECT .+ FROM orders WHERE order_uid = \$1`).
		WithArgs("order-1").
		WillReturnRows(order1Rows)

	delivery1Rows := mock.NewRows([]string{
		"name", "phone", "zip", "city", "address", "region", "email",
	}).AddRow("User 1", "+111", "123", "City1", "Addr1", "Reg1", "email1@test.com")

	mock.ExpectQuery(`SELECT .+ FROM delivery WHERE order_uid = \$1`).
		WithArgs("order-1").
		WillReturnRows(delivery1Rows)

	payment1Rows := mock.NewRows([]string{
		"transaction", "request_id", "currency", "provider", "amount",
		"payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee",
	}).AddRow("trans1", "", "USD", "wbpay", 100, int64(123456), "bank1", 10, 90, 0)

	mock.ExpectQuery(`SELECT .+ FROM payment WHERE order_uid = \$1`).
		WithArgs("order-1").
		WillReturnRows(payment1Rows)

	items1Rows := mock.NewRows([]string{
		"chrt_id", "track_number", "price", "rid", "name", "sale",
		"size", "total_price", "nm_id", "brand", "status",
	})

	mock.ExpectQuery(`SELECT .+ FROM items WHERE order_uid = \$1 ORDER BY id`).
		WithArgs("order-1").
		WillReturnRows(items1Rows)

	order2Rows := mock.NewRows([]string{
		"order_uid", "track_number", "entry", "locale", "internal_signature",
		"customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard",
	}).AddRow("order-2", "TRACK2", "WBIL", "ru", "", "cust2", "meest", "9", int64(99), now, "1")

	mock.ExpectQuery(`SELECT .+ FROM orders WHERE order_uid = \$1`).
		WithArgs("order-2").
		WillReturnRows(order2Rows)

	delivery2Rows := mock.NewRows([]string{
		"name", "phone", "zip", "city", "address", "region", "email",
	}).AddRow("User 2", "+222", "456", "City2", "Addr2", "Reg2", "email2@test.com")

	mock.ExpectQuery(`SELECT .+ FROM delivery WHERE order_uid = \$1`).
		WithArgs("order-2").
		WillReturnRows(delivery2Rows)

	payment2Rows := mock.NewRows([]string{
		"transaction", "request_id", "currency", "provider", "amount",
		"payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee",
	}).AddRow("trans2", "", "USD", "wbpay", 200, int64(123456), "bank2", 20, 180, 0)

	mock.ExpectQuery(`SELECT .+ FROM payment WHERE order_uid = \$1`).
		WithArgs("order-2").
		WillReturnRows(payment2Rows)

	items2Rows := mock.NewRows([]string{
		"chrt_id", "track_number", "price", "rid", "name", "sale",
		"size", "total_price", "nm_id", "brand", "status",
	})

	mock.ExpectQuery(`SELECT .+ FROM items WHERE order_uid = \$1 ORDER BY id`).
		WithArgs("order-2").
		WillReturnRows(items2Rows)

	orders, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, orders, 2)

	assert.Equal(t, "order-1", orders[0].OrderUID)
	assert.Equal(t, "User 1", orders[0].Delivery.Name)

	assert.Equal(t, "order-2", orders[1].OrderUID)
	assert.Equal(t, "User 2", orders[1].Delivery.Name)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresOrderRepo_GetAll_Empty(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &PostgresOrderRepo{pool: mock}

	idRows := mock.NewRows([]string{"order_uid"})

	mock.ExpectQuery(`SELECT order_uid FROM orders ORDER BY date_created DESC`).
		WillReturnRows(idRows)

	orders, err := repo.GetAll()
	require.NoError(t, err)
	assert.Empty(t, orders)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresOrderRepo_Ping_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &PostgresOrderRepo{pool: mock}

	mock.ExpectPing()

	err = repo.Ping()
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresOrderRepo_Ping_Error(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &PostgresOrderRepo{pool: mock}

	mock.ExpectPing().WillReturnError(errors.New("connection error"))

	err = repo.Ping()
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
