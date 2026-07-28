BINARY_NAME=server
BUILD_DIR=bin
CMD_DIR=./cmd/server

.PHONY: all build run test check clean format help
all: format test build

build:
	@echo "Building the binary..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

run:
	@echo "Running the program..."
	@go run $(CMD_DIR)

test:
	@echo "Running tests..."
	go test -v -race ./internal/...

check:
	@echo "Checking all the project..."
	go build $(CMD_DIR)
	go test -v -race ./internal/...

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)

format:
	@echo "Formatting code..."
	go fmt ./...

help:
	@echo "Available commands:"
	@echo "  make build   - Build the go binary"
	@echo "  make run     - Run the application (go run)"
	@echo "  make test    - Run tests"
	@echo "  make check   - Try to build the project and run tests"
	@echo "  make clean   - Remove build artifacts"
	@echo "  make format  - Format go code"
