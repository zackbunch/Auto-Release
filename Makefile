.PHONY: all build run clean tidy

# Default target
all: build

# Build the syac binary
build:
	@echo "Building syac binary..."
	go build -o syac .

local:
	go run main.go

# Run the syac binary (e.g., show help)
run: build
	@echo "Running syac..."
	./syac --help

# Clean up build artifacts
clean:
	@echo "Cleaning up build artifacts..."
	rm -f syac

# Run go mod tidy
tidy:
	@echo "Running go mod tidy..."
	go mod tidy
