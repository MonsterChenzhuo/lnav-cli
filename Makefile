# Copyright (c) 2026 lnav-cli authors
# SPDX-License-Identifier: MIT

GO       ?= go
BINARY   ?= lnav-cli
MODULE   := github.com/MonsterChenzhuo/lnav-cli
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo DEV)
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE     ?= $(shell date -u +%Y-%m-%d)
PREFIX   ?= /usr/local

LDFLAGS  := -s -w \
	-X $(MODULE)/internal/build.Version=$(VERSION) \
	-X $(MODULE)/internal/build.Commit=$(COMMIT) \
	-X $(MODULE)/internal/build.Date=$(DATE)

.PHONY: help build install uninstall clean fmt vet lint \
        unit-test coverage e2e-dryrun e2e-live skills-check test release-snapshot

help: ## show this help
	@awk 'BEGIN{FS=":.*##"; printf "Targets:\n"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## build the binary with version metadata
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .

install: build ## install to $(PREFIX)/bin
	install -d $(PREFIX)/bin
	install -m755 $(BINARY) $(PREFIX)/bin/$(BINARY)
	@echo "OK: $(PREFIX)/bin/$(BINARY) ($(VERSION))"

uninstall: ## remove from $(PREFIX)/bin
	rm -f $(PREFIX)/bin/$(BINARY)

clean: ## remove build artifacts
	rm -f $(BINARY)
	rm -rf dist/ coverage.out

fmt: ## apply gofmt
	gofmt -l -w .

vet: ## go vet
	$(GO) vet ./...

lint: ## run golangci-lint
	$(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run

unit-test: ## unit + cmd tests with race detector
	$(GO) test -race -count=1 ./internal/... ./cmd/...

coverage: ## unit tests with coverage profile
	$(GO) test -race -count=1 -coverprofile=coverage.out ./internal/... ./cmd/...
	$(GO) tool cover -func=coverage.out | tail -n 1

e2e-dryrun: ## dry-run E2E tests (no lnav required)
	$(GO) test -race -count=1 ./tests/e2e/dryrun/...

e2e-live: ## live E2E tests (lnav must be installed)
	$(GO) test -count=1 ./tests/e2e/live/...

skills-check: ## validate skills/*/SKILL.md frontmatter
	bash scripts/skills-check.sh

test: vet unit-test e2e-dryrun skills-check ## full local CI gate

release-snapshot: ## dry-run goreleaser build locally
	$(GO) run github.com/goreleaser/goreleaser/v2@latest build --snapshot --clean --single-target
