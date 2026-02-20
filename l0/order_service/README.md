# Order Service

Микросервис для работы с заказами. Получает данные из Kafka, сохраняет в PostgreSQL, кэширует в Redis, отдает JSON по HTTP по запросу 

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

# Технологии
- Go 1.24
- PostgreSQL 15
- Redis 7
- Kafka (KRaft mode)
- Gin
- pgx
- go-redis
- segmentio/kafka-go

TODO:
    TESTS
    MOVE INTERFACES
    CLEAN
