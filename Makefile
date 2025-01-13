CLUSTER_NAME := axial

dev-env:
	@$(MAKE) cluster-exists || ( \
		echo "Creating kind cluster '$(CLUSTER_NAME)'..."; \
		kind create cluster --name $(CLUSTER_NAME); \
	)
	@echo "Run \`tilt up\` to start the development environment."

clean:
	# Clean up other dangling resources
	@echo "Cleaning up any leftover containers or volumes..."
	@docker ps -a --filter "name=$(CLUSTER_NAME)" --format '{{.ID}}' | xargs -r docker rm -f
	@docker volume ls --filter "dangling=true" --format '{{.Name}}' | xargs -r docker volume rm
	@docker system prune --force --volumes >/dev/null


.PHONY: dev-env create-cluster delete-cluster cluster-exists clean

# Check if the cluster exists
cluster-exists:
	@kind get clusters | grep -q "^$(CLUSTER_NAME)$$" || false
