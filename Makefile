VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

VAULT_ADDR ?= http://127.0.0.1:8200
VAULT_TOKEN ?= dev-root

.PHONY: all build test test-integration fuzz cover lint fmt vuln secrets check release-snapshot dev-up dev-down clean

all: lint test build

build: ## Build the CLI and the API into bin/
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/secureenv ./cmd/secureenv
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/secureenv-api ./cmd/secureenv-api

test: ## Run unit and end-to-end tests
	go test -race ./...

test-integration: ## Run the Vault contract, OpenAPI conformance and end-to-end tests against a real Vault (see dev-up)
	VAULT_ADDR=$(VAULT_ADDR) VAULT_TOKEN=$(VAULT_TOKEN) go test -race -tags integration ./internal/vaultstore/ ./internal/e2e/

fuzz: ## Fuzz the .env parser
	go test -run '^$$' -fuzz FuzzFormatParseRoundTrip -fuzztime 60s ./internal/dotenv/

cover: ## Write an HTML coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format the code
	golangci-lint fmt ./...

vuln: ## Report known vulnerabilities reachable from the code
	go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...

secrets: ## Scan the git history for secrets (requires gitleaks)
	gitleaks git --config .gitleaks.toml --redact --no-banner .

check: lint test vuln secrets ## Run the same checks as the CI, except Docker and Vault

release-snapshot: ## Build release archives for every platform into dist/ (requires goreleaser)
	goreleaser release --snapshot --clean

dev-up: ## Start a local Vault dev server and the API
	docker compose up -d --build --wait

dev-down: ## Stop the local stack
	docker compose down

clean:
	rm -rf bin dist coverage.out
