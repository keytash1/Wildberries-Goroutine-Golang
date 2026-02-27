.PHONY: build up down logs produce lint test help

build:
	docker-compose build

up:
	docker-compose up -d
	@echo "Services started"
	@echo "   Service: http://localhost:8081"
	@echo "   Prometheus: http://localhost:9090"
	@echo "   Jaeger: http://localhost:16686"

down:
	docker-compose down -v

logs:
	docker-compose logs -f

produce:
	docker-compose run --rm producer-test

lint:
	golangci-lint run --fix ./...

test-unit:
	go test -v ./...

test-integration:
	go test -v ./test/integration/... -timeout 180s

test: 
	test-unit test-integration

help:
	@echo "Available commands:"
	@echo "  make build           - build Docker images"
	@echo "  make up              - start all services"
	@echo "  make down            - stop all services"
	@echo "  make logs            - view all logs"
	@echo "  make produce         - send test orders"
	@echo "  make lint            - run linter with fix"
	@echo "  make test-unit       - run unit tests"
	@echo "  make test-integration - run integration tests"
	@echo "  make test            - run all tests"