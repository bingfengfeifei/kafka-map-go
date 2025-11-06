# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Kafka-Map Go is a Kafka visualization and management tool with a Go backend and React frontend. The application is compiled into a single binary with embedded frontend assets using Go's `embed` package.

## Build and Development Commands

### Backend Development
```bash
# Run in development mode (no frontend build required)
make dev

# Build full application (includes frontend build)
make build

# Build optimized production binary
make build-prod

# Run tests
make test

# Format Go code
make fmt

# Lint Go code (requires golangci-lint)
make lint
```

### Frontend Development
```bash
cd web
npm install --legacy-peer-deps
npm run dev          # Development server
npm run build        # Production build
```

### Docker
```bash
make docker-build    # Build Docker image
make docker-run      # Run container with volume mount
```

## Architecture

### Layered Architecture Pattern

The backend follows a clean layered architecture:

1. **Controller Layer** (`internal/controller/`) - HTTP handlers, request/response handling
2. **Service Layer** (`internal/service/`) - Business logic, orchestration
3. **Repository Layer** (`internal/repository/`) - Data access, database operations
4. **Model Layer** (`internal/model/`) - Database entities (GORM models)
5. **DTO Layer** (`internal/dto/`) - Data transfer objects for API contracts

### Dependency Injection

The application uses manual dependency injection initialized in `cmd/server/main.go`:
- Repositories depend on database connection
- Services depend on repositories and utilities
- Controllers depend on services
- All dependencies are created at startup and passed through constructors

### Kafka Client Management

**KafkaClientManager** (`internal/util/kafka_client.go`) is a critical component:
- Maintains a cache of Kafka admin clients per cluster (map[uint]sarama.ClusterAdmin)
- Thread-safe with RWMutex for concurrent access
- Creates clients on-demand and reuses them for performance
- Handles multiple security protocols: PLAINTEXT, SASL_PLAINTEXT, SASL_SSL, SSL
- Supports SASL mechanisms: PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
- Factory methods: GetAdminClient(), CreateConsumer(), CreateClient(), CreateProducer()

### Authentication System

Token-based authentication with in-memory cache:
- **TokenCache** (`internal/util/token_cache.go`) - In-memory token storage with expiration
- **AuthMiddleware** (`internal/middleware/auth.go`) - Validates tokens on protected routes
- Default credentials: admin/admin (configurable in `config/config.yaml`)
- Tokens expire after 2 hours (configurable via `cache.token_expiration`)

### Frontend Embedding

The React frontend is embedded into the Go binary:
- Build process: `npm run build` creates `web/dist/` with `index.html` and `assets/`
- Assets are copied to `cmd/server/web/` before Go build
- `//go:embed web/index.html web/assets` directive embeds files into binary
- Custom routing in main.go serves assets with proper MIME types
- SPA routing: all non-API routes serve index.html for client-side routing

### Database

- SQLite with GORM ORM
- Database file location: `data/kafka-map.db` (configurable)
- Auto-migration on startup via `database.InitDB()`
- Two main entities: User and Cluster

## Key Implementation Details

### Cluster Configuration

Clusters are stored in SQLite and contain:
- Connection details (servers, security protocol)
- SASL credentials (username, password, mechanism)
- Each cluster gets a cached Kafka admin client in KafkaClientManager

### API Structure

All API routes are under `/api` prefix:
- Public: `/api/login`
- Protected: All other routes require authentication token in header
- RESTful design with standard HTTP methods
- Query parameter `clusterId` required for most Kafka operations

### Error Handling

Services return Go errors, controllers translate to HTTP responses:
- Use `fmt.Errorf()` with `%w` for error wrapping
- Controllers return JSON with `code` and `message` fields
- Gin's `c.JSON()` for consistent response format

### Frontend Build Integration

The Makefile target `build-frontend`:
1. Runs `npm install --legacy-peer-deps` (required for dependency resolution)
2. Runs `npm run build` (Vite builds to `web/dist/`)
3. Creates `cmd/server/web/` directory
4. Copies `web/index.html` and `web/assets/` to `cmd/server/web/`
5. Go build then embeds these files

**Important**: When modifying frontend, always rebuild with `make build` or manually run `make build-frontend` before `go build`.

## Configuration

Edit `config/config.yaml`:
- `server.port` - HTTP server port (default: 8080)
- `database.path` - SQLite database file path
- `default.username/password` - Initial admin credentials
- `cache.token_expiration` - Token TTL in seconds
- `cache.max_tokens` - Maximum concurrent tokens

## Testing Kafka Connections

When adding Kafka cluster support:
1. Cluster model must include security protocol and SASL mechanism
2. KafkaClientManager.buildConfig() handles protocol configuration
3. Test with different security protocols to ensure proper TLS/SASL setup
4. Note: TLS uses `InsecureSkipVerify: true` - should be configurable for production

## Common Patterns

### Adding a New API Endpoint

1. Define DTO in `internal/dto/` if needed
2. Add method to appropriate service in `internal/service/`
3. Add handler to controller in `internal/controller/`
4. Register route in `cmd/server/main.go` (public or protected group)

### Working with Kafka

Always use KafkaClientManager methods:
- For admin operations: `GetAdminClient()` (cached)
- For consuming: `CreateConsumer()` (new instance)
- For producing: `CreateProducer()` (new instance)
- Remember to close non-cached clients after use

### Frontend API Calls

Frontend uses axios configured in `src/common/request.js`:
- Base URL from environment variable
- Automatic token injection from localStorage
- Centralized error handling
