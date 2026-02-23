# Order Service

Микросервис для работы с заказами. Получает данные из Kafka, сохраняет в PostgreSQL, кэширует в Redis, отдает JSON по HTTP запросу.

# 1. Запустить инфраструктуру
docker-compose up --build

# 2. Отправить тестовые заказы
docker-compose up producer-test

ID созданного заказа можно найти в logs

# 3. Ручки
http://localhost:8081

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/` | Веб-интерфейс |
| GET | `/order/:id` | Получить заказ по ID (JSON) |
| GET | `/health` | Проверка здоровья |

# 4. Запустить юнит-тесты
go test -v ./...

### Компоненты
- **HTTP Server** - принимает запросы от клиентов (`/order/:id`, `/`, `/health`)
- **Kafka Consumer** - получает сообщения о новых заказах из Kafka
- **Service Layer** - бизнес-логика, валидация, работа с кэшем и БД
- **Repository** - работа с PostgreSQL (транзакции)
- **Cache** - кэширование в Redis (с функцией восстановления их БД)
- **PostgreSQL** - основное хранилище (таблицы: orders, delivery, payment, items)
- **Redis** - кэш для быстрого доступа (TTL 24 часа)
- **Kafka** - очередь сообщений (KRaft mode)


## Структура проекта
order-service/
├── cmd/ # точки входа
│ ├── producer/ # тестовый продюсер
│ │ └── main.go
│ └── server/ # основной сервис
│ └── main.go
├── internal/ # внутренний код
│ ├── cache/ # работа с Redis
│ │ ├── order_cache.go # интерфейс кэша
│ │ ├── redis.go # реализация Redis
│ │ └── redis_test.go # тесты (miniredis)
│ ├── config/ # конфигурация
│ │ └── config.go
│ ├── domain/ # модели данных
│ │ └── models.go
│ ├── handler/ # HTTP обработчики
│ │ ├── order_handler.go # API /order/:id
│ │ ├── web_handler.go # веб-интерфейс
│ │ └── web/ # встроенные файлы
│ │ ├── static/ # CSS/JS
│ │ │ ├── css/
│ │ │ │ └── style.css
│ │ │ └── js/
│ │ │ └── app.js
│ │ └── templates/ # HTML
│ │ └── index.html
│ ├── kafka/ # работа с Kafka
│ │ ├── consumer.go
│ │ ├── consumer_test.go
│ │ └── producer.go
│ ├── repository/ # работа с PostgreSQL
│ │ ├── order_repo.go # интерфейс репозитория
│ │ ├── postgres.go # реализация
│ │ └── postgres_test.go # тесты (pgxmock)
│ └── service/ # бизнес-логика
│ ├── order_service.go
│ └── order_service_test.go # тесты (моки)
├── migrations/ # SQL миграции
│ ├── 001_init_schema.down.sql
│ └── 001_init_schema.up.sql
├── docker-compose.yml # инфраструктура
├── Dockerfile.server # сборка сервиса
├── Dockerfile.producer # сборка продюсера
├── go.mod
├── go.sum
├── model.json # тестовые данные
└── README.md

## Технологии

- **Go 1.24** - основной язык
- **Gin** - HTTP фреймворк
- **pgx** - драйвер PostgreSQL
- **go-redis** - клиент Redis
- **segmentio/kafka-go** - клиент Kafka
- **PostgreSQL 15** - основная БД
- **Redis 7** - кэш
- **Kafka 4.2 (KRaft)** - очередь сообщений
- **Docker** - контейнеризация
- **testify** - тестирование (assert, mock)
- **pgxmock** - моки для БД
- **miniredis** - моки для Redis


## Механизмы работы

### 1. Получение сообщений из Kafka (at-least-once)
Consumer использует ручной коммит offset'ов:
```go
msg, _ := reader.FetchMessage(ctx)
if err := processMessage(msg); err != nil {
    continue // не коммитим при ошибке
}
reader.CommitMessages(ctx, msg) // коммитим только при успехе
```

### 2. Сохранение в БД (транзакции)
Все операции (orders, delivery, payment, items) в одной транзакции

При ошибке любого INSERT делаем откат (Rollback)

ON CONFLICT DO NOTHING для идемпотентности

### 3. Кэширование
При сохранении заказ сразу пишется в Redis (TTL 24 часа)

При получении сначала проверяется кэш (Get)

При старте сервиса кэш восстанавливается из БД (RestoreCache)

## Что тестируется

| Пакет | Тесты | Инструменты |
|-------|-------|-------------|
| **repository** | `Save`, `GetByID`, `GetAll`, `Ping`, обработка ошибок | `pgxmock` |
| **cache** | `Set`, `Get`, `Restore`, TTL, ошибки подключения | `miniredis` |
| **service** | `GetOrder` (cache hit/miss), `SaveOrder` (валидация), `RestoreCache` | `testify/mock` |
| **kafka** | `processMessage` (успех, ошибки), `Start`/`Stop` | `testify/mock` |

TODO:
    Оптимизировать GetAll in repo
    Исправить Graceful Shtd
    Переместить интерфейсы
    golint
    Убрать тестовый продюсер из докера
