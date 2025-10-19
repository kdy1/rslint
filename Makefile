.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: ## Run all tests
	go test ./internal/...

.PHONY: test-typescript-estree
test-typescript-estree: ## Run typescript-estree module tests
	go test ./internal/typescript-estree/...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: test-coverage-typescript-estree
test-coverage-typescript-estree: ## Run typescript-estree tests with coverage
	go test -coverprofile=coverage-typescript-estree.out ./internal/typescript-estree/...
	go tool cover -html=coverage-typescript-estree.out -o coverage-typescript-estree.html

.PHONY: lint
lint: ## Run linters
	golangci-lint run ./cmd/... ./internal/...

.PHONY: lint-typescript-estree
lint-typescript-estree: ## Run linters on typescript-estree module
	golangci-lint run ./internal/typescript-estree/...

.PHONY: fmt
fmt: ## Format Go code
	golangci-lint fmt ./cmd/... ./internal/...

.PHONY: build
build: ## Build the project
	go build ./cmd/...

.PHONY: tidy
tidy: ## Run go mod tidy on all modules
	go mod tidy
	cd internal/typescript-estree && go mod tidy

.PHONY: clean
clean: ## Clean build artifacts
	rm -f coverage.out coverage.html coverage-typescript-estree.out coverage-typescript-estree.html
	go clean -cache
