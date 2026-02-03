CLUSTER_NAME := axial

# Development environment with PostgreSQL and Go app
# Rebuild Go binary when any Go files (or go.mod/go.sum) change
GO_SOURCES := $(shell find src -type f -name '*.go')

src/axial: src/frontend/dist $(GO_SOURCES) src/go.mod src/go.sum
	@echo "Building Go application..."
	@cd src && go build -o axial

run: src/axial
	@echo "Starting PostgreSQL and pgAdmin..."
	@cd docker && docker compose up -d
	@echo "Starting application..."
	@cd src && sudo ./axial || true

dev: dev-clean src/axial
	@echo "Starting PostgreSQL and pgAdmin..."
	@cd docker && docker compose up -d
	@echo "Starting application in development mode..."
	@cd src && sudo ./axial

dev-clean:
	@rm -rf src/axial

.PHONY: deps
deps:
  @echo "Installing dependencies..."
	@go install github.com/air-verse/air@latest

clean:
	# Clean up other dangling resources
	@echo "Cleaning web/dist..."
	@rm -rf web/dist
	@echo "Cleaning web/node_modules..."
	@rm -rf web/node_modules
	@echo "Cleaning src/frontend/dist..."
	@rm -rf src/frontend/dist
	@echo "Cleaning src/axial..."
	@rm -rf src/axial
	@echo "Cleaning up any leftover containers and volumes..."
	@cd docker && docker compose down -v
	@docker volume rm axial_postgres_data 2>/dev/null || true
	@docker ps -a --filter "name=$(CLUSTER_NAME)" --format '{{.ID}}' | xargs -r docker rm -f
	@docker volume ls --filter "dangling=true" --format '{{.Name}}' | xargs -r docker volume rm
	@docker system prune --force --volumes >/dev/null

.PHONY: dev-env create-cluster delete-cluster cluster-exists clean run dev

# Check if the cluster exists
cluster-exists:
	@kind get clusters | grep -q "^$(CLUSTER_NAME)$$" || false

web/node_modules:
	@echo "Installing node modules..."
	@cd web && npm install

src/frontend/dist: web/node_modules
	cd web && npm install && npm run build