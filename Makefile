CLUSTER_NAME := axial

# Development environment with PostgreSQL and Go app
run:
	@echo "Starting PostgreSQL and pgAdmin..."
	@cd docker && docker compose up -d
	@echo "Building Go application..."
	@cd src && go build -o axial
	@echo "Starting application..."
	@cd src && sudo ./axial || true
	@echo "Cleaning up..."
	@cd docker && docker compose down

dev-env:
	@$(MAKE) cluster-exists || ( \
		echo "Creating kind cluster '$(CLUSTER_NAME)'..."; \
		kind create cluster --name $(CLUSTER_NAME); \
	)
	@echo "Run \`tilt up\` to start the development environment."

clean:
	# Clean up other dangling resources
	@echo "Cleaning up any leftover containers and volumes..."
	@cd docker && docker compose down -v
	@docker volume rm axial_postgres_data 2>/dev/null || true
	@docker ps -a --filter "name=$(CLUSTER_NAME)" --format '{{.ID}}' | xargs -r docker rm -f
	@docker volume ls --filter "dangling=true" --format '{{.Name}}' | xargs -r docker volume rm
	@docker system prune --force --volumes >/dev/null


.PHONY: dev-env create-cluster delete-cluster cluster-exists clean run

# Check if the cluster exists
cluster-exists:
	@kind get clusters | grep -q "^$(CLUSTER_NAME)$$" || false
