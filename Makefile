.PHONY: help build install test testacc fmt vet clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \\033[36m%-18s\\033[0m %s\\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the provider
build: ## Build the Terraform provider
	CGO_ENABLED=0 go build -o terraform-provider-quismon

# Install the provider locally for development
install: build ## Install provider locally for development
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/quismon/quismon/1.0.0/linux_amd64/
	cp terraform-provider-quismon ~/.terraform.d/plugins/registry.terraform.io/quismon/quismon/1.0.0/linux_amd64/

# Run unit tests
test: ## Run unit tests
	CGO_ENABLED=0 go test ./... -v -cover

# Run acceptance tests (requires API key)
testacc: ## Run acceptance tests (requires API key)
	CGO_ENABLED=0 TF_ACC=1 go test ./internal/provider -v -timeout 120m

# Run all tests with detailed output
test-all: ## Run all tests with detailed output
	@./scripts/test.sh all

# Run matrix tests (all variations)
test-matrix: ## Run matrix tests (all variations)
	@./scripts/test.sh matrix

# Run integration tests
test-integration: ## Run integration tests
	@./scripts/test.sh integration

# Run quick smoke tests
test-quick: ## Run quick smoke tests
	@./scripts/test.sh quick

# Format code
fmt: ## Format code with gofmt
	go fmt ./...

# Run go vet
vet: ## Run go vet linter
	go vet ./...

# Clean build artifacts
clean: ## Clean build artifacts
	rm -f terraform-provider-quismon
	rm -rf ~/.terraform.d/plugins/registry.terraform.io/quismon/

# Generate documentation
docs: ## Generate Terraform provider documentation
	go generate ./...

# Run all checks before committing
check: fmt vet test ## Run all pre-commit checks

.DEFAULT_GOAL := help
