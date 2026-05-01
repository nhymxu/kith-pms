.DEFAULT_GOAL := list

# Variables
APP_NAME=kith-pms

# Insert a comment starting with '##' after a target, and it will be printed by 'make' and 'make help'
.PHONY: list
list: ## list Makefile targets
	@echo "Available commands: \n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: tools
tools: ## Install pinned dev tools (sqlc, templ, tailwindcss standalone binary)
	bash scripts/install-tools.sh

.PHONY: sqlc
sqlc: ## Generate sqlc query code
	sqlc generate -f internal/db/sqlc.yaml

.PHONY: templ
templ: ## Generate templ components
	rm -f internal/web/templates/templates_stub.go
	templ generate ./internal/web/...

.PHONY: tailwind
tailwind: ## Build Tailwind CSS output
	bin/tailwindcss -i internal/web/templates/styles.css -o internal/web/static/app.css --minify

.PHONY: assets
assets: sqlc templ tailwind ## Regenerate all generated assets (sqlc + templ + tailwind)

.PHONY: build
build: assets ## Build the binary (CGO_ENABLED=0)
	CGO_ENABLED=0 go build -o bin/$(APP_NAME) ./cmd

.PHONY: migrate
migrate: ## Apply database migrations
	./bin/$(APP_NAME) migrate up

.PHONY: dev
dev: ## Run development servers (templ watch + tailwind watch + go run)
	templ generate --watch ./internal/web/... & \
	bin/tailwindcss -i internal/web/templates/styles.css -o internal/web/static/app.css --watch & \
	CGO_ENABLED=0 go run ./cmd api

.PHONY: deps
deps: ## Install Go dependencies
	go mod download
	go mod tidy

.PHONY: check-fmt
check-fmt: ## Ensure code is formatted
	gofmt -l -d .
	test -z "$$(gofmt -l .)"

.PHONY: fmt
fmt: ## Run go fmt against code
	go fmt ./...

.PHONY: tests
tests: ## Run all tests
	go test -race -v -tags integration $(GO_TEST_FLAGS)

.PHONY: test-coverage
test-coverage:
	go test -coverprofile=cover.out -shuffle on ./...

.PHONY: cover
cover:
	go tool cover -html=cover.out

.PHONY: benchmark
benchmark:
	go test -bench=.

.PHONY: lint
lint: ## Run linter
	./scripts/lint.sh

## tidy: format code and tidy modfile
.PHONY: tidy
tidy: fmt
	go mod tidy -v

.PHONY: find-cgo-pkg
find-cgo-pkg: ## Identify which packages use CGO
	./scripts/find-cgo-pkg.sh

.PHONY: check-duplicate-code
check-duplicate-code: ## Identify duplicate code
	go install github.com/boyter/dcd@latest
	dcd

.PHONY: vuln-check
vuln-check: ## Scan source code for vulnerabilities
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck

.PHONY: vuln-check-bin
vuln-check-bin: ## Scan binary for vulnerabilities
	go build -o test_binary
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck -mode binary -show verbose test_binary

.PHONY: gosec
gosec: ## Analyse Go source for security issues
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec ./...
