.PHONY: help build up down restart logs clean test test-verbose generate migrate db-status db-reset

help:
	@echo "Easy Tactics - Available commands:"
	@echo ""
	@echo "  make build        - Build Docker images"
	@echo "  make up           - Start all services"
	@echo "  make down         - Stop all services"
	@echo "  make restart      - Restart all services"
	@echo "  make logs         - View logs"
	@echo "  make clean        - Remove containers and volumes"
	@echo ""
	@echo "  make test         - Run tests"
	@echo "  make test-verbose - Run tests with verbose output"
	@echo ""
	@echo "  make generate     - Generate protobuf code"
	@echo ""
	@echo "  make migrate      - Run database migrations"
	@echo "  make db-status    - Show migration status"
	@echo "  make db-reset     - Reset database (warning: deletes data)"

build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

restart:
	docker-compose restart

logs:
	docker-compose logs -f

clean:
	docker-compose down -v
	rm -f data/fighters.db

test:
	cd api && go test ./...

test-verbose:
	cd api && go test ./... -v

migrate:
	docker-compose exec api goose -dir /app/migrations sqlite3 /data/fighters.db up

db-status:
	docker-compose exec api goose -dir /app/migrations sqlite3 /data/fighters.db status

db-reset:
	@echo "WARNING: This will delete all data!"
	@read -p "Are you sure? (yes/no) " confirm && [ "$$confirm" = "yes" ] || exit 1
	docker-compose down -v
	rm -f data/fighters.db
	docker-compose up -d

generate:
	@echo "Generating protobuf code..."
	cd api && go generate ./...