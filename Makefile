.PHONY: build run

# Build with gojson for faster JSON (high-load)
build:
	go build -tags=go_json -o bin/gophermart ./cmd/gophermart

run:
	go run -tags=go_json ./cmd/gophermart
