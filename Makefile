.PHONY: build build-arm clean build-frontend release

# Application version
VERSION ?= 1.0.0

# Detect host architecture for UPX selection
HOST_ARCH := $(shell uname -m)
ifeq ($(HOST_ARCH),x86_64)
	UPX_BIN := upx/x86/upx
else ifeq ($(HOST_ARCH),aarch64)
	UPX_BIN := upx/arm/upx
else ifeq ($(HOST_ARCH),arm64)
	UPX_BIN := upx/arm/upx
else
	UPX_BIN := upx/x86/upx
endif

# Build the Go application
build: build-frontend
	@echo "Building Kafka-Map Go for Linux AMD64..."
	CGO_ENABLED=0 go build -ldflags="-s -w"  -trimpath  -o kafka-map-go ./cmd/server
	@echo "Compressing binary with UPX (using $(UPX_BIN))..."
	chmod +x $(UPX_BIN)
	./$(UPX_BIN) --best --lzma kafka-map-go

# Build optimized binary for Linux ARM64
build-arm: build-frontend
	@echo "Building Kafka-Map Go for Linux ARM64..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o kafka-map-go ./cmd/server
	@echo "Compressing binary with UPX (using $(UPX_BIN))..."
	chmod +x $(UPX_BIN)
	./$(UPX_BIN) --best --lzma kafka-map-go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f kafka-map-go
	rm -rf data/
	rm -rf web/dist

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

# Create release packages
release: clean
	@echo "Creating release packages for version $(VERSION)..."

	# Build AMD64
	@echo "Building AMD64 binary..."
	@$(MAKE) build
	@mkdir -p release/kafka-map-go-v$(VERSION)-linux-amd64
	@cp kafka-map-go release/kafka-map-go-v$(VERSION)-linux-amd64/
	@cp -r config release/kafka-map-go-v$(VERSION)-linux-amd64/
	@cd release && tar -czf kafka-map-go-v$(VERSION)-linux-amd64.tar.gz kafka-map-go-v$(VERSION)-linux-amd64
	@echo "Created release/kafka-map-go-v$(VERSION)-linux-amd64.tar.gz"

	# Clean binary for ARM64 build
	@rm -f kafka-map-go

	# Build ARM64
	@echo "Building ARM64 binary..."
	@$(MAKE) build-arm
	@mkdir -p release/kafka-map-go-v$(VERSION)-linux-arm64
	@cp kafka-map-go release/kafka-map-go-v$(VERSION)-linux-arm64/
	@cp -r config release/kafka-map-go-v$(VERSION)-linux-arm64/
	@cd release && tar -czf kafka-map-go-v$(VERSION)-linux-arm64.tar.gz kafka-map-go-v$(VERSION)-linux-arm64
	@echo "Created release/kafka-map-go-v$(VERSION)-linux-arm64.tar.gz"

	# Cleanup temporary directories
	@rm -rf release/kafka-map-go-v$(VERSION)-linux-amd64
	@rm -rf release/kafka-map-go-v$(VERSION)-linux-arm64

	@echo ""
	@echo "Release packages created successfully:"
	@ls -lh release/*.tar.gz
