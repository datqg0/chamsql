.PHONY: up down run sqlc new_migration migrate build clean tidy

# Load environment variables from .env
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Docker
up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

# Development
run:
	air

tidy:
	go mod tidy

build:
	go build -o dist/api ./cmd/app

clean:
	rm -rf dist/

# SQLC
sqlc:
	sqlc generate

# Goose migrations
new_migration:
	@read -p "Enter migration name: " name; \
	goose -dir sql/schema create $$name sql

migrate:
	@echo "Migrating DB: $(DATABASE_URI)"
	goose -dir sql/schema postgres "$(DATABASE_URI)" up

migrate_down:
	goose -dir sql/schema postgres "$(DATABASE_URI)" down

migrate_status:
	goose -dir sql/schema postgres "$(DATABASE_URI)" status

# All-in-one
dev: up tidy run
