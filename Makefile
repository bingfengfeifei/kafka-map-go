.PHONY: build run test clean install-deps build-frontend docker-build docker-run

# Build the Go application
build: build-frontend
	@echo "Building Kafka-Map Go..."
	CGO_ENABLED=0 go build -o kafka-map-go ./cmd/server

# Build with optimizations for production
build-prod: build-frontend
	@echo "Building Kafka-Map Go for production..."
	CGO_ENABLED=0 go build -ldflags="-s -w"  -trimpath  -o kafka-map-go ./cmd/server

# Run the application
run: build
	@echo "Running Kafka-Map Go..."
	./kafka-map-go

# Run in development mode
dev:
	@echo "Running in development mode..."
	go run ./cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f kafka-map-go
	rm -rf data/
	rm -rf web/dist web/node_modules

# Install Go dependencies
install-deps:
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy

# Build frontend
build-frontend:
	@echo "Building frontend..."
	cd web && npm install --legacy-peer-deps && npm run build
	mkdir -p cmd/server/web
	rm -rf cmd/server/web/assets
	cp -rf web/dist/index.html web/dist/assets cmd/server/web/

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t kafka-map-go:latest .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -v $(PWD)/data:/app/data kafka-map-go:latest

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  build-prod     - Build for production with optimizations"
	@echo "  run            - Build and run the application"
	@echo "  dev            - Run in development mode"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  install-deps   - Install Go dependencies"
	@echo "  build-frontend - Build frontend assets"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
