# Billing Hotspot — developer convenience targets.
# Run `make help` for the list.

.DEFAULT_GOAL := help
SHELL := /bin/bash

# ── Docker (per-service stacks) ───────────────────────────────────────────────
# Each service has its OWN docker-compose.yml so it can be started/stopped and
# maintained independently. They communicate over a shared external network.
.PHONY: net
net: ## Create the shared docker network (run once)
	docker network inspect hotspot-net >/dev/null 2>&1 || docker network create hotspot-net

# radius-api stack: mariadb + radius-api + freeradius
.PHONY: radius-up radius-down radius-logs
radius-up: net ## Start the radius stack (mariadb + radius-api + freeradius)
	@test -f radius-api/.env || cp radius-api/.env.example radius-api/.env
	cd radius-api && docker compose up -d --build
radius-down: ## Stop the radius stack
	cd radius-api && docker compose down
radius-logs: ## Tail radius stack logs
	cd radius-api && docker compose logs -f

# backend stack: postgres + backend
.PHONY: backend-up backend-down backend-logs
backend-up: net ## Start the backend stack (postgres + backend)
	@test -f backend/.env || cp backend/.env.example backend/.env
	cd backend && docker compose up -d --build
backend-down: ## Stop the backend stack
	cd backend && docker compose down
backend-logs: ## Tail backend stack logs
	cd backend && docker compose logs -f

# frontend stack: nginx-served SPA
.PHONY: frontend-up frontend-down frontend-logs
frontend-up: ## Start the frontend stack
	@test -f frontend/.env || cp frontend/.env.example frontend/.env
	cd frontend && docker compose up -d --build
frontend-down: ## Stop the frontend stack
	cd frontend && docker compose down
frontend-logs: ## Tail frontend stack logs
	cd frontend && docker compose logs -f

# Convenience: bring everything up in dependency order (still separate stacks).
.PHONY: up down
up: radius-up backend-up frontend-up ## Start all stacks (in order)
down: frontend-down backend-down radius-down ## Stop all stacks

.PHONY: clean
clean: ## Stop all stacks and DELETE their volumes (data loss!)
	-cd frontend && docker compose down -v
	-cd backend && docker compose down -v
	-cd radius-api && docker compose down -v

# ── Backend ──────────────────────────────────────────────────────────────────
.PHONY: backend-run
backend-run: ## Run the billing backend locally
	cd backend && go run ./cmd/api

.PHONY: backend-build
backend-build: ## Compile the billing backend
	cd backend && go build ./...

.PHONY: backend-swagger
backend-swagger: ## Regenerate backend Swagger docs (needs `swag`)
	cd backend && swag init -g cmd/api/main.go -o docs

# ── radius-api ───────────────────────────────────────────────────────────────
.PHONY: radius-run
radius-run: ## Run the radius-api locally
	cd radius-api && go run ./cmd/api

.PHONY: radius-build
radius-build: ## Compile the radius-api
	cd radius-api && go build ./...

.PHONY: radius-swagger
radius-swagger: ## Regenerate radius-api Swagger docs (needs `swag`)
	cd radius-api && swag init -g cmd/api/main.go -o docs

# ── Frontend ─────────────────────────────────────────────────────────────────
.PHONY: frontend-dev
frontend-dev: ## Run the frontend dev server (Vite)
	cd frontend && pnpm install && pnpm dev

.PHONY: frontend-build
frontend-build: ## Build the frontend for production
	cd frontend && pnpm install && pnpm build

# ── Tooling ──────────────────────────────────────────────────────────────────
.PHONY: install-swag
install-swag: ## Install the swaggo CLI used to regenerate API docs
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
