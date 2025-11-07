.PHONY: build build-arm clean build-frontend

# Application version
VERSION ?= 1.0.0

# Build the Go application
build: build-frontend
	@echo "Building Kafka-Map Go for Linux AMD64..."
	CGO_ENABLED=0 go build -ldflags="-s -w"  -trimpath  -o kafka-map-go ./cmd/server
	@echo "Compressing binary with UPX..."
	chmod +x upx/x86/upx
	./upx/x86/upx --best --lzma kafka-map-go

# Build optimized binary for Linux ARM64
build-arm: build-frontend
	@echo "Building Kafka-Map Go for Linux ARM64..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o kafka-map-go-arm64 ./cmd/server
	@echo "Compressing binary with UPX..."
	chmod +x upx/arm/upx
	./upx/arm/upx --best --lzma kafka-map-go-arm64

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f kafka-map-go
	rm -rf data/
	rm -rf web/dist web/node_modules

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

# Build Docker image
docker:
	@echo "Building Docker image with version $(VERSION)..."
	docker build --build-arg APP_VERSION=$(VERSION) -t kafka-map-go:$(VERSION) -t kafka-map-go:latest .
