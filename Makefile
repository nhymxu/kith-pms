.DEFAULT_GOAL := list

# Variables
APP_NAME=go-boilerplate

# Insert a comment starting with '##' after a target, and it will be printed by 'make' and 'make help'
.PHONY: help
help: ## list Makefile targets
	@echo "Available commands: \n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: deps
deps: ## Install dependencies
	go mod download
	go mod tidy

.PHONY: check-fmt
check-fmt: ## Ensure code is formatted
	gofmt -l -d . 	# For the sake of debugging
	test -z "$$(gofmt -l .)"

.PHONY: fmt
fmt: ## Run go fmt against code
	go fmt ./...

.PHONY: tests
tests: ## Run all tests and requires a running rabbitmq-server. Use GO_TEST_FLAGS to add extra flags to go test
	go test -race -v -tags integration $(GO_TEST_FLAGS)

.PHONY: test-coverage
test:
	go test -coverprofile=cover.out -shuffle on ./...

.PHONY: cover
cover:
	go tool cover -html=cover.out

.PHONY: benchmark
benchmark:
	go test -bench=.

.PHONY: lint
check:
	./scripts/lint.sh

## tidy: format code and tidy modfile
.PHONY: tidy
tidy: fmt
	go mod tidy -v

.PHONY: find-cgo-pkg
find-cgo-pkg:  ## identify which package on project using CGO
	./scripts/find-cgo-pkg.sh

.PHONY: check-duplicate-code
check-duplicate-code: ## identify duplicate code inside a project
	go install github.com/boyter/dcd@latest
	dcd

.PHONY: vuln-check
vuln-check:  ## Scanning source code for vulnerabilities
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck

.PHONY: vuln-check-bin
vuln-check-bin:  ## Scanning binary for vulnerabilities
	go build -o test_binary
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck -mode binary -show verbose test_binary

.PHONY: gosec
gosec:  ## analyses Go source code to look for common programming mistakes that can lead to security problems.
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec ./...
