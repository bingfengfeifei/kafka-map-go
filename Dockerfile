# Build stage for frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app/web
COPY web/package*.json ./
RUN npm install --legacy-peer-deps

COPY web/ ./
RUN npm run build

# Build stage for Go application
FROM golang:1.21-alpine AS go-builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy frontend build artifacts
COPY --from=frontend-builder /app/web/dist/index.html ./cmd/server/web/
COPY --from=frontend-builder /app/web/dist/assets ./cmd/server/web/assets

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o kafka-map-go ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

# Copy binary and config
COPY --from=go-builder /app/kafka-map-go .
COPY --from=go-builder /app/config ./config

# Create data directory
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Run the application
CMD ["./kafka-map-go"]
