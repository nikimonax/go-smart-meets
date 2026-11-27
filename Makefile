include .env

MODULE := $(shell go list -m)
APP_VERSION := v0.1.0

BIN_DIR  := ./bin
OUT_DIR  := ./out
DOCS_DIR := ./docs


COV_FILE := $(OUT_DIR)/coverage.out
COV_FILE_HTML := $(OUT_DIR)/coverage.html


GOOSE_DBSTRING := "postgres://$(DATABASE_USER):$(DATABASE_PASS)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=disable"
GOOSE_MIGRATION_DIR := internal/impl/database/migrations
GOOSE_DRIVER := postgres


EXTRA_ARGS += $(ARGS)


all: build


.SUFFIXES:

.PHONY: .FORCE
.FORCE:

.env: .env.example
	cp $< $@

# ========== Command ========== #

.PHONY: lint
lint: _check_golangci_lint_cmd _check_protolint_cmd
	golangci-lint run
	protolint .

.PHONY: fmt format
fmt format: _check_golangci_lint_cmd
	golangci-lint fmt

.PHONY: test
test: | $(OUT_DIR)
	go test $(EXTRA_ARGS) $$(go list ./... | grep -v internal/testing | grep -v internal/gen)

.PHONY: cover
cover: $(COV_FILE)
	sed -i '/internal\/testing/d' $(COV_FILE) && \
	sed -i '/internal\/gen/d' $(COV_FILE) && \
	go tool cover -func=$(COV_FILE)

.PHONY: cover-html
cover-html: $(COV_FILE_HTML)

$(COV_FILE): EXTRA_ARGS += -coverpkg=./... -coverprofile=$(COV_FILE)
$(COV_FILE): test | $(OUT_DIR)

$(COV_FILE_HTML): $(COV_FILE) | $(OUT_DIR)
	go tool cover -html=$< -o $@

.PHONY: migrate-up migrate-down migrate-status
migrate-up migrate-down migrate-status: migrate-%: _check_goose_cmd
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
	GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) \
	GOOSE_DRIVER=$(GOOSE_DRIVER) \
	goose $* $(EXTRA_ARGS)

.PHONY: river-goose-generate
river-goose-generate:
	go run ./tools/river-goose-generate -dir $(GOOSE_MIGRATION_DIR)

.PHONY: up down logs
up: EXTRA_ARGS+=-d
logs: EXTRA_ARGS+=-f
up down logs:
	docker compose $@ $(EXTRA_ARGS)

.PHONY: clean
clean:
	rm -rf $(BIN_DIR) $(OUT_DIR)
	go clean -testcache

$(BIN_DIR) $(OUT_DIR) $(DOCS_DIR):
	mkdir -p $@

# ========== Compile ========== #

.PHONY: build
build: $(BIN_DIR)/smart-meets-server

$(BIN_DIR)/smart-meets-server: $(BIN_DIR)/%: cmd/% .FORCE
	go build -o $@ ./$<

.PHONY: run-server
run-server: $(BIN_DIR)/smart-meets-server
	./$< $(EXTRA_ARGS)

# ========== checks ========== #

.PHONY: _check_golangci_lint_cmd
_check_golangci_lint_cmd:
	@if ! command -v golangci-lint $&> /dev/null; then \
		echo "[ERROR]: command 'golangci-lint' not found"; \
		echo "install using \"go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest\""; \
		exit 1; \
	fi

.PHONY: _check_goose_cmd
_check_goose_cmd:
	@if ! command -v goose $&> /dev/null; then \
		echo "[ERROR]: command 'goose' not found"; \
		echo "install using \"go install github.com/pressly/goose/v3/cmd/goose@latest\""; \
		exit 1; \
	fi


.PHONY: _check_protolint_cmd
_check_protolint_cmd:
	@if ! command -v protolint $&> /dev/null; then \
		echo "[ERROR]: command 'protolint' not found"; \
		echo "install using \"go install github.com/yoheimuta/protolint/cmd/protolint@latest\""; \
		exit 1; \
	fi
