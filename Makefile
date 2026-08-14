SHELL := /bin/bash
.DEFAULT_GOAL := help

BACKEND_DIR := backend/recap
BACKEND_GO_MOD := $(BACKEND_DIR)/go.mod
BIN_DIR := $(CURDIR)/bin
APP_BIN := $(BIN_DIR)/recap
SEED_YEAR ?= 2026
SEED_VALUE ?= 20260807

GOLANGCI_LINT_VERSION := v2.12.2
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint
GOLANGCI_CONFIG := $(CURDIR)/.golangci.yml
GOOSE_VERSION := v3.27.1
GOOSE := $(BIN_DIR)/goose
GOOSE_TABLE := public.goose_db_version
OGEN_VERSION := v1.23.0
OGEN := $(BIN_DIR)/ogen
MOCKERY_VERSION := v3.7.0
MOCKERY := $(BIN_DIR)/mockery

ENV_FILE := $(CURDIR)/.env
MIGRATIONS_DIR := $(CURDIR)/backend/recap/migrations/migrations
OPENAPI_SPEC := $(CURDIR)/backend/recap/api/recap/v1/openapi.yaml
GENERATED_API_DIR := $(CURDIR)/backend/recap/generated/recapapi
POSTGRES_DSN = host=$${POSTGRES_BIND_HOST:-127.0.0.1} \
	port=$${POSTGRES_EXTERNAL_PORT:-5432} \
	user=$${POSTGRES_USER} \
	password=$${POSTGRES_PASSWORD} \
	dbname=$${POSTGRES_DB} \
	sslmode=$${POSTGRES_SSL_MODE:-disable}

.PHONY: help tools require-backend require-env lint-config format lint vet test test-race \
	test-integration test-e2e tidy tidy-check generate generate-api generate-mocks check up down logs \
	build run db-up compose-config ps logs-recap migrate-up migrate-down migrate-status \
	seed seed-reset seed-dry-run

help:
	@echo "Available commands:"
	@echo "  make tools         Install development tools"
	@echo "  make format        Format Go code"
	@echo "  make lint          Run golangci-lint"
	@echo "  make vet           Run go vet"
	@echo "  make test          Run unit tests"
	@echo "  make test-race     Run unit tests with the race detector"
	@echo "  make test-integration Run integration tests when present"
	@echo "  make test-e2e      Run the backend HTTP journey against PostgreSQL"
	@echo "  make tidy          Synchronize Go dependencies"
	@echo "  make tidy-check    Check whether go.mod and go.sum are tidy"
	@echo "  make generate      Generate code from project contracts"
	@echo "  make generate-mocks Generate mocks for unit tests"
	@echo "  make check         Run all required Go checks"
	@echo "  make build         Build recap service locally"
	@echo "  make run           Run recap locally with PostgreSQL in Docker"
	@echo "  make db-up         Start PostgreSQL only"
	@echo "  make up            Build and start the complete application stack"
	@echo "  make down          Stop the application stack"
	@echo "  make compose-config Validate Docker Compose configuration"
	@echo "  make ps            Show Compose service status"
	@echo "  make logs          Follow all application logs"
	@echo "  make logs-recap    Follow recap service logs"
	@echo "  make migrate-up    Apply database migrations"
	@echo "  make migrate-down  Roll back the latest migration"
	@echo "  make migrate-status Show database migration status"
	@echo "  make seed          Generate and load deterministic demo data"
	@echo "  make seed-reset    Replace existing demo data"
	@echo "  make seed-dry-run  Preview generated demo data without PostgreSQL"

$(GOLANGCI_LINT):
	mkdir -p $(BIN_DIR)
	GOBIN=$(BIN_DIR) go install \
		github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

$(GOOSE):
	mkdir -p $(BIN_DIR)
	GOBIN=$(BIN_DIR) go install \
		github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)

$(OGEN):
	mkdir -p $(BIN_DIR)
	GOBIN=$(BIN_DIR) go install \
		github.com/ogen-go/ogen/cmd/ogen@$(OGEN_VERSION)

$(MOCKERY):
	mkdir -p $(BIN_DIR)
	GOBIN=$(BIN_DIR) go install \
		github.com/vektra/mockery/v3@$(MOCKERY_VERSION)

tools: $(GOLANGCI_LINT) $(GOOSE) $(OGEN) $(MOCKERY)

require-backend:
	@test -f $(BACKEND_GO_MOD) || { \
		echo "Backend module not found: $(BACKEND_GO_MOD)"; \
		echo "Create it before running Go checks."; \
		exit 1; \
	}

require-env:
	@test -f $(ENV_FILE) || { \
		echo "Environment file not found: $(ENV_FILE)"; \
		echo "Run: cp .env.example .env"; \
		exit 1; \
	}
	@set -a; source $(ENV_FILE); set +a; \
	for variable in POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB; do \
		if [ -z "$${!variable}" ]; then \
			echo "Required environment variable is empty: $$variable"; \
			exit 1; \
		fi; \
	done

lint-config: require-backend $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) config verify --config=$(GOLANGCI_CONFIG)

format: require-backend $(GOLANGCI_LINT)
	cd $(BACKEND_DIR) && $(GOLANGCI_LINT) fmt --config=$(GOLANGCI_CONFIG)

lint: require-backend $(GOLANGCI_LINT)
	cd $(BACKEND_DIR) && $(GOLANGCI_LINT) run --config=$(GOLANGCI_CONFIG) ./...

vet: require-backend
	cd $(BACKEND_DIR) && go vet ./...

test: require-backend generate
	cd $(BACKEND_DIR) && go test ./...

test-race: require-backend generate
	cd $(BACKEND_DIR) && go test -race -count=1 ./...

test-integration: require-backend
	@set -euo pipefail; \
	tests="$$(find $(BACKEND_DIR) -type f -name '*_test.go' \
		-exec grep -l '^//go:build integration' {} + 2>/dev/null || true)"; \
	if [ -z "$$tests" ]; then \
		echo "No integration tests found; skipping."; \
		exit 0; \
	fi; \
	cd $(BACKEND_DIR) && go test -race -count=1 -timeout=5m -tags=integration ./tests/integration/...

test-e2e: require-backend
	@set -euo pipefail; \
	tests="$$(find $(BACKEND_DIR) -type f -name '*_test.go' \
		-exec grep -l '^//go:build e2e' {} + 2>/dev/null || true)"; \
	if [ -z "$$tests" ]; then \
		echo "No e2e tests found; skipping."; \
		exit 0; \
	fi; \
	cd $(BACKEND_DIR) && go test -race -count=1 -timeout=5m -tags=e2e ./tests/e2e/...

tidy: require-backend
	cd $(BACKEND_DIR) && go mod tidy

tidy-check: require-backend
	cd $(BACKEND_DIR) && go mod tidy -diff

generate: generate-mocks

generate-api: require-backend $(OGEN)
	$(OGEN) --target $(GENERATED_API_DIR) --package recapapi --clean $(OPENAPI_SPEC)

generate-mocks: generate-api $(MOCKERY)
	cd $(BACKEND_DIR) && $(MOCKERY) --config .mockery.yml

check: generate
	$(MAKE) lint-config lint vet tidy-check test-race test-integration test-e2e

build: generate-api
	mkdir -p $(BIN_DIR)
	cd $(BACKEND_DIR) && \
		CGO_ENABLED=0 go build -trimpath -o $(APP_BIN) ./cmd/app

db-up: require-env
	docker compose up -d --wait postgres

run: require-env generate-api db-up
	@set -a; source $(ENV_FILE); set +a; \
	cd $(BACKEND_DIR) && go run ./cmd/app

seed: require-env db-up
	docker compose run --rm --build seed \
		--year=$(SEED_YEAR) \
		--seed=$(SEED_VALUE)

seed-reset: require-env db-up
	docker compose run --rm --build seed \
		--year=$(SEED_YEAR) \
		--seed=$(SEED_VALUE) \
		--reset

up: require-env generate-api
	docker compose up -d --build --wait

down:
	docker compose down

compose-config: require-env
	docker compose config --quiet

ps:
	docker compose ps

logs:
	docker compose logs -f

logs-recap:
	docker compose logs -f recap

migrate-up: require-env $(GOOSE)
	@set -a; source $(ENV_FILE); set +a; \
	$(GOOSE) -table $(GOOSE_TABLE) -dir $(MIGRATIONS_DIR) postgres \
		"$(POSTGRES_DSN)" \
		up

migrate-down: require-env $(GOOSE)
	@set -a; source $(ENV_FILE); set +a; \
	$(GOOSE) -table $(GOOSE_TABLE) -dir $(MIGRATIONS_DIR) postgres \
		"$(POSTGRES_DSN)" \
		down

migrate-status: require-env $(GOOSE)
	@set -a; source $(ENV_FILE); set +a; \
	$(GOOSE) -table $(GOOSE_TABLE) -dir $(MIGRATIONS_DIR) postgres \
		"$(POSTGRES_DSN)" \
		status
