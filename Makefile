.PHONY: build run

# go_json: Gin uses goccy/go-json instead of encoding/json for binding/rendering
build:
	go build -tags=go_json -o bin/gophermart ./cmd/gophermart

run:
	go run -tags=go_json ./cmd/gophermart
