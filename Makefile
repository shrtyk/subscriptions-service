include .env

.PHONY: dto/generate docker/up docker/down app/run db/start db/stop psql \
		migrations/up migrations/up-by-one migrations/down migrations/down-all migrations/status

MIGRATIONS_DIR=./migrations
UNIT_TESTS_PKGS := $(shell go list ./... | grep -v /mocks | grep -v /gen | grep -v /dto | grep -v /cmd | grep -v /ports)

# Setup enviroment to run
setup:
	@cp .env_example .env

# Build containers and start them in background
docker/up:
	@docker compose up -d --build

# Stop and remove containers with their volumes
docker/down:
	@docker compose down --volumes

# Run app
app/run:
	@trap 'docker-compose down subscriptions' EXIT; \
	docker compose up --build subscriptions

# Run db with migrations in background
db/start:
	@docker compose up -d --build postgres goose
	@docker compose logs goose

# Stop db
db/stop:
	@docker compose down --volumes postgres goose

# Open psql in postgres container
psql:
	@docker compose exec postgres psql -U $(PG_USER) -d $(PG_DBNAME)

# Creating migrations
migrations/new:
	@if [ -z "$(NAME)" ]; then \
		echo "NAME is not set. Usage: make migrations/new NAME=your_migration_name"; \
		exit 1; \
	fi
	@goose -dir $(MIGRATIONS_DIR) create $(NAME) sql

# Migrate the DB to the most recent version available
migrations/up:
	@docker compose run --rm goose up

# Migrate the DB up by 1
migrations/up-by-one:
	@docker compose run --rm goose up-by-one

# Roll back the version by 1
migrations/down:
	@docker compose run --rm goose down

# Roll back all migrations
migrations/down-all:
	@docker compose run --rm goose down-to 0

# Dump the migration status for the current DB
migrations/status:
	@docker compose run --rm goose status

# Generate DTOs
dto/generate:
	@go generate ./internal/api/http/dto/dto.go

# Generate mocks
mocks/generate:
	@go generate ./internal/core/ports/...

# Run unit tests
unit-tests/run:
	@mkdir -p coverage
	@go test -v -race \
    -coverprofile=coverage.out -covermode=atomic ${UNIT_TESTS_PKGS}

# Run linter
linter/run:
	@golangci-lint run ./...
