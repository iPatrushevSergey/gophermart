.PHONY: build run test test-unit test-contract test-integration test-e2e test-all cover

APP_DIR := app

# go_json: Gin uses goccy/go-json instead of encoding/json for binding/rendering
build:
	cd $(APP_DIR) && go build -tags=go_json -o bin/gophermart ./cmd/gophermart

run:
	cd $(APP_DIR) && go run -tags=go_json ./cmd/gophermart

test:
	$(MAKE) test-unit

test-unit:
	cd $(APP_DIR) && go test ./internal/... ./cmd/...

test-contract:
	cd $(APP_DIR) && go test ./tests/contract/...

test-integration:
	cd $(APP_DIR) && go test -tags integration ./internal/gophermart/adapters/repository/postgres/...

test-e2e:
	cd $(APP_DIR) && go test -tags integration ./tests/e2e/...

test-all: test-unit test-contract test-integration test-e2e

cover:
	cd $(APP_DIR) && go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out
