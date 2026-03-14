.PHONY: dev build test lint fmt vet clean

BINARY_NAME=dreams
BUILD_DIR=./build
CMD_PATH=./cmd/main.go

dev:
	go run $(CMD_PATH)

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)

test:
	go test ./...

lint:
	go vet ./...
	golangci-lint run 2>/dev/null || echo "golangci-lint not installed, skipping"

fmt:
	gofmt -s -w .

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
	rm -f ./var/dreams.db
