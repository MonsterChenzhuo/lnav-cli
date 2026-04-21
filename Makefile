.PHONY: build unit-test lint e2e-dryrun e2e-live skills-check test fmt vet

GO ?= go
BINARY ?= lnav-cli

build:
	$(GO) build -o $(BINARY) .

unit-test:
	$(GO) test -race -count=1 ./internal/... ./cmd/...

e2e-dryrun:
	$(GO) test -race -count=1 ./tests/e2e/dryrun/...

e2e-live:
	$(GO) test -count=1 ./tests/e2e/live/...

skills-check:
	bash scripts/skills-check.sh

fmt:
	gofmt -l -w .

vet:
	$(GO) vet ./...

lint:
	$(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run

test: unit-test e2e-dryrun skills-check
