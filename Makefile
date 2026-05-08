SHELL := /bin/sh

COMPOSE ?= docker compose -f deploy/docker-compose.yml --env-file .env
BACKEND_DIR := backend
FRONTEND_DIR := frontend
SMOKE_SCRIPT := scripts/smoke.sh

.PHONY: dev test smoke docker-build clean ensure-env

ensure-env:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env from .env.example"; \
	fi

dev: ensure-env
	$(COMPOSE) up --build

test:
	@if [ -f "$(BACKEND_DIR)/go.mod" ]; then \
		echo "Running backend tests"; \
		(cd "$(BACKEND_DIR)" && go test ./...); \
	else \
		echo "Skipping backend tests: $(BACKEND_DIR)/go.mod is not present yet"; \
	fi
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		echo "Running frontend tests"; \
		if [ -f "$(FRONTEND_DIR)/package.json" ]; then \
			(cd "$(FRONTEND_DIR)" && npm test -- --run); \
		else \
			echo "Skipping frontend tests: package.json is not present yet"; \
		fi; \
	else \
		echo "Skipping frontend tests: $(FRONTEND_DIR)/ is not present yet"; \
	fi

smoke:
	@if [ ! -x "$(SMOKE_SCRIPT)" ]; then chmod +x "$(SMOKE_SCRIPT)"; fi
	@"$(SMOKE_SCRIPT)"

docker-build: ensure-env
	@if [ ! -f "$(BACKEND_DIR)/Dockerfile" ] || [ ! -f "$(FRONTEND_DIR)/Dockerfile" ]; then \
		echo "Cannot build Docker images until backend/Dockerfile and frontend/Dockerfile exist"; \
		exit 1; \
	fi
	$(COMPOSE) build

clean:
	@if [ -f .env ]; then \
		$(COMPOSE) down --remove-orphans; \
	else \
		docker compose -f deploy/docker-compose.yml down --remove-orphans; \
	fi
	rm -rf data/uploads/*
