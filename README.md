# Kafka-Map Go

A Kafka visualization and management tool written in Go, converted from the original Java Spring Boot version.

## Features

- **Multi-cluster Management** - Add, remove, and manage multiple Kafka clusters
- **Cluster Monitoring** - View cluster status, partition counts, replica counts, storage size, offsets
- **Topic Management** - Create, delete, expand partitions, view configurations
- **Broker Monitoring** - View broker status and configurations
- **Consumer Group Management** - View, delete consumer groups, monitor lag
- **Offset Management** - Reset consumer group offsets (beginning, end, or specific offset)
- **Message Operations** - View, search, and send messages to topics
- **Authentication** - Token-based authentication with password management
- **Embedded Frontend** - React frontend embedded in single binary

## Technology Stack

### Backend
- **Gin** - Web framework
- **GORM** - ORM for database operations
- **Sarama** - Kafka client library
- **SQLite** - Embedded database
- **BCrypt** - Password encryption

### Frontend
- **React** - UI framework
- **Vite** - Build tool
- Embedded using Go's `embed` package

## Prerequisites

- Go 1.21 or higher
- Node.js 18+ and npm (for building frontend)
- GCC (for SQLite CGO support)

## Quick Start

### Using Make

```bash
# Install dependencies and build
make install-deps
make build

# Run the application
make run

# Or run in development mode
make dev
```

### Manual Build

```bash
# Install Go dependencies
go mod download

# Build frontend
cd web
npm install --legacy-peer-deps
npm run build
cd ..

# Copy frontend assets
mkdir -p cmd/server/web
cp -r web/index.html web/assets cmd/server/web/

# Build Go application
go build -o kafka-map-go ./cmd/server

# Run
./kafka-map-go
```

### Using Docker

```bash
# Build Docker image
docker build -t kafka-map-go:latest .

# Run container
docker run -p 8080:8080 -v $(pwd)/data:/app/data kafka-map-go:latest
```

## Configuration

Edit `config/config.yaml`:

```yaml
server:
  port: 8080

database:
  path: data/kafka-map.db

default:
  username: admin
  password: admin

cache:
  token_expiration: 7200  # 2 hours in seconds
  max_tokens: 100
```

## Default Credentials

- **Username**: admin
- **Password**: admin

**Important**: Change the default password after first login!

## API Endpoints

### Authentication
- `POST /api/login` - User login
- `POST /api/logout` - User logout
- `GET /api/info` - Get current user info
- `POST /api/change-password` - Change password

### Clusters
- `GET /api/clusters` - List all clusters
- `GET /api/clusters/:id` - Get cluster details
- `POST /api/clusters` - Create cluster
- `PUT /api/clusters/:id` - Update cluster
- `DELETE /api/clusters/:id` - Delete cluster

### Brokers
- `GET /api/brokers?clusterId=:id` - List brokers
- `GET /api/brokers/:id/configs?clusterId=:id` - Get broker configs
- `PUT /api/brokers/:id/configs?clusterId=:id` - Update broker configs

### Topics
- `GET /api/topics?clusterId=:id` - List topics
- `GET /api/topics/names?clusterId=:id` - Get topic names
- `GET /api/topics/:topic?clusterId=:id` - Get topic details
- `POST /api/topics?clusterId=:id` - Create topic
- `POST /api/topics/batch-delete?clusterId=:id` - Delete topics
- `POST /api/topics/:topic/partitions?clusterId=:id` - Expand partitions
- `PUT /api/topics/:topic/configs?clusterId=:id` - Update topic configs
- `GET /api/topics/:topic/data?clusterId=:id` - Get messages
- `POST /api/topics/:topic/data?clusterId=:id` - Send message

### Consumer Groups
- `GET /api/consumerGroups?clusterId=:id` - List consumer groups
- `GET /api/consumerGroups/:groupId?clusterId=:id` - Get group details
- `DELETE /api/consumerGroups/:groupId?clusterId=:id` - Delete group
- `GET /api/topics/:topic/consumerGroups?clusterId=:id` - Get groups by topic
- `GET /api/topics/:topic/consumerGroups/:groupId/offset?clusterId=:id` - Get offsets
- `PUT /api/topics/:topic/consumerGroups/:groupId/offset?clusterId=:id` - Reset offsets

## Project Structure

```
kafka-map-go/
├── cmd/
│   └── server/
│       ├── main.go              # Application entry point
│       └── web/                 # Embedded frontend assets
├── internal/
│   ├── config/                  # Configuration management
│   ├── controller/              # HTTP handlers
│   ├── dto/                     # Data transfer objects
│   ├── middleware/              # Middleware (auth, CORS)
│   ├── model/                   # Database models
│   ├── repository/              # Data access layer
│   ├── service/                 # Business logic
│   └── util/                    # Utilities
├── pkg/
│   └── database/                # Database initialization
├── web/                         # Frontend source code
├── config/
│   └── config.yaml              # Configuration file
├── Dockerfile
├── Makefile
└── README.md
```

## Development

### Running Tests

```bash
make test
```

### Code Formatting

```bash
make fmt
```

### Linting

```bash
make lint
```

### Cleaning Build Artifacts

```bash
make clean
```

## Security Features

- **Token-based Authentication** - 2-hour token expiration
- **BCrypt Password Hashing** - Secure password storage
- **CORS Support** - Configurable cross-origin requests
- **SASL/SCRAM Support** - Secure Kafka authentication (PLAIN, SCRAM-SHA-256, SCRAM-SHA-512)
- **SSL/TLS Support** - Encrypted Kafka connections

## Supported Kafka Security Protocols

- PLAINTEXT
- SASL_PLAINTEXT
- SASL_SSL
- SSL

## Supported SASL Mechanisms

- PLAIN
- SCRAM-SHA-256
- SCRAM-SHA-512

## Differences from Java Version

### Not Included
- **Delay Message System** - The 18-level delay message feature is not implemented in this version

### Improvements
- **Single Binary** - Frontend embedded in Go binary
- **Smaller Footprint** - No JVM required
- **Faster Startup** - Native binary execution
- **Simpler Deployment** - Single executable file

## Building for Production

```bash
# Build optimized binary
make build-prod

# The binary will be created as 'kafka-map-go'
# Deploy with config directory and run
./kafka-map-go
```

## Troubleshooting

### CGO Errors
If you encounter CGO-related errors, ensure you have GCC installed:
```bash
# Ubuntu/Debian
apt-get install build-essential

# Alpine
apk add gcc musl-dev

# macOS
xcode-select --install
```

### Frontend Build Issues
If frontend build fails, try:
```bash
cd web
rm -rf node_modules package-lock.json
npm install --legacy-peer-deps
npm run build
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is converted from the original Java version of Kafka-Map.

## Acknowledgments

- Original Java version: [kafka-map](https://github.com/typesafe/kafka-map)
- Built with [Gin](https://github.com/gin-gonic/gin)
- Kafka client: [Sarama](https://github.com/IBM/sarama)
- ORM: [GORM](https://gorm.io)
