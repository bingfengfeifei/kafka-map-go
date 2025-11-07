# Runtime stage - expects kafka-map-go binary from make build
ARG ALPINE_VERSION=3.19
ARG APP_VERSION=1.0.0
FROM alpine:${ALPINE_VERSION}

WORKDIR /app

# Copy binary and config from build context
COPY kafka-map-go .
COPY config ./config

# Create data directory
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Run the application
CMD ["./kafka-map-go"]
