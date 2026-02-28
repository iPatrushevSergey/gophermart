.PHONY: build run test test-integration cover

APP_DIR := app

# go_json: Gin uses goccy/go-json instead of encoding/json for binding/rendering
build:
	cd $(APP_DIR) && go build -tags=go_json -o bin/gophermart ./cmd/gophermart

run:
	cd $(APP_DIR) && go run -tags=go_json ./cmd/gophermart

test:
	cd $(APP_DIR) && go test ./...

test-integration:
	cd $(APP_DIR) && go test -tags integration ./...

cover:
	cd $(APP_DIR) && go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out
