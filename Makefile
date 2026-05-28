.DEFAULT_GOAL := list

# Variables
APP_NAME=kith-pms

# Insert a comment starting with '##' after a target, and it will be printed by 'make' and 'make help'
.PHONY: list
list: ## list Makefile targets
	@echo "Available commands: \n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: tools
tools: ## Install pinned dev tools (sqlc)
	bash scripts/install-tools.sh

.PHONY: sqlc
sqlc: ## Generate sqlc query code
	sqlc generate -f internal/db/sqlc.yaml

.PHONY: web
web: ## Build the React SPA and copy output into internal/api/spa/public
	cd web && pnpm install --frozen-lockfile && pnpm build
	rm -rf internal/api/spa/public
	mkdir -p internal/api/spa/public
	touch internal/api/spa/public/.gitignore
	cp -R web/dist/. internal/api/spa/public

.PHONY: assets
assets: sqlc web ## Regenerate all generated assets (sqlc + SPA build)

.PHONY: build
build: web ## Build the binary (CGO_ENABLED=0); runs pnpm build first
	CGO_ENABLED=0 go build -o bin/$(APP_NAME) ./cmd

.PHONY: clean
clean: ## Remove build artefacts (web/dist and internal/api/spa/public)
	rm -rf web/dist internal/api/spa/public

.PHONY: migrate
migrate: ## Apply database migrations
	./bin/$(APP_NAME) migrate up

.PHONY: dev
dev: ## Run dev servers (pnpm dev + go run serve) — SPA proxies /v1 to :8000
	@trap 'kill 0' EXIT INT TERM; \
	pnpm --dir web dev 2>&1 | sed 's/^/[web] /' & \
	CGO_ENABLED=0 go run ./cmd serve 2>&1 | sed 's/^/[api] /' & \
	wait

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
	go test -race -v -tags integration $(GO_TEST_FLAGS) ./...

.PHONY: test-coverage
test-coverage:
	go test -coverprofile=cover.out -shuffle on ./...

.PHONY: cover
cover:
	go tool cover -html=cover.out

.PHONY: benchmark
benchmark:
	go test -bench=.

.PHONY: lint-go
lint-go: ## Run golang linter
	echo "----- Run linting for backend using golangci-lint"
	./scripts/lint.sh

.PHONY: lint-biome
lint-biome:
	echo "----- Run linting for frontend using biome"
	cd web && pnpm biome check && cd ..

.PHONY: lint-tsc
lint-tsc:
	echo "----- Run linting for frontend using tsc"
	cd web && pnpm tsc --noEmit && cd ..

.PHONY: lint
lint: lint-go lint-biome lint-tsc

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
