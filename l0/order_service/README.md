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

| Путь | Назначение |
|------|------------|
| `cmd/producer/` | Тестовый продюсер для Kafka |
| `cmd/server/` | Основной сервис |
| `internal/cache/` | Работа с Redis (кэш) |
| `internal/config/` | Конфигурация |
| `internal/domain/` | Модели данных |
| `internal/handler/` | HTTP обработчики (API + веб) |
| `internal/handler/web/` | Встроенные HTML/CSS/JS |
| `internal/kafka/` | Consumer/Producer для Kafka |
| `internal/repository/` | Работа с PostgreSQL |
| `internal/service/` | Бизнес-логика и валидация |
| `migrations/` | SQL миграции |
| `docker-compose.yml` | Инфраструктура (PostgreSQL, Redis, Kafka) |
| `Dockerfile.*` | Docker-образы |
| `model.json` | Тестовые данные |

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

## Переменные окружения

Скопируйте файл `.env.example` в `.env` и отредактируйте при необходимости:
```bash
cp .env.example .env
```
Файл .env находится в gitignore:
```gitignore
# Environment variables
.env
```

1 Отсутствуют интеграционные тесты
2 Трейсы и метрики должны присутствовать в продакшен-реди коде
об изменениях в ридми тесты мейк

make
trace
metrics
integrtests
ivalid producer
dlq cachelogs
dockertestconteainer
linter vscode