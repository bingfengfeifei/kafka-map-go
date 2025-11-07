# Kafka Map Go

English | [简体中文](README_CN.md)

A Kafka visualization and management tool written in Go, converted from the original Java Spring Boot version.

**Key Advantages:**
- **Single Binary** - No dependencies required, runs standalone
- **Lightweight** - Binary size < 10MB with embedded frontend
- **Fast Deployment** - Just download and run, no installation needed

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
- **SQLite** - Embedded database (no CGO dependency)
- **BCrypt** - Password encryption

### Frontend
- **React** - UI framework
- **Vite** - Build tool
- Embedded using Go's `embed` package

## Prerequisites

- Go 1.21 or higher
- Node.js 18+ and npm (for building frontend)

## Quick Start

### Using Make

```bash
# Build the application (includes frontend build)
make build

# Build for ARM64
make build-arm
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
# Build Docker image (optionally specify version)
make docker VERSION=1.0.0

# Or use docker directly
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

### Environment Variables

You can override any of the YAML settings (or provide new defaults) via the following environment variables:

| Variable | Description |
| --- | --- |
| `KAFKA_MAP_SERVER_PORT` | Override `server.port`. |
| `KAFKA_MAP_DATABASE_PATH` | Override `database.path`. |
| `DEFAULT_USERNAME` / `KAFKA_MAP_DEFAULT_USERNAME` | Override the initial admin username. |
| `DEFAULT_PASSWORD` / `KAFKA_MAP_DEFAULT_PASSWORD` | Override the initial admin password. |
| `KAFKA_MAP_CACHE_TOKEN_EXPIRATION` | Override `cache.token_expiration` (seconds). |
| `KAFKA_MAP_CACHE_MAX_TOKENS` | Override `cache.max_tokens`. |
| `DEFAULT_CLUSTER_NAME` / `KAFKA_MAP_BOOTSTRAP_NAME` | Name of a cluster to auto-create at startup. |
| `DEFAULT_CLUSTER_SERVERS` / `KAFKA_MAP_BOOTSTRAP_SERVERS` | Comma-separated broker list for the bootstrap cluster. |
| `DEFAULT_CLUSTER_SECURITY_PROTOCOL` / `KAFKA_MAP_BOOTSTRAP_SECURITY_PROTOCOL` | Optional security protocol (defaults to `PLAINTEXT`). |
| `DEFAULT_CLUSTER_SASL_MECHANISM` / `KAFKA_MAP_BOOTSTRAP_SASL_MECHANISM` | Optional SASL mechanism. |
| `DEFAULT_CLUSTER_SASL_USERNAME` / `KAFKA_MAP_BOOTSTRAP_SASL_USERNAME` | SASL username (if required). |
| `DEFAULT_CLUSTER_SASL_PASSWORD` / `KAFKA_MAP_BOOTSTRAP_SASL_PASSWORD` | SASL password (if required). |
| `DEFAULT_CLUSTER_AUTH_USERNAME` / `KAFKA_MAP_BOOTSTRAP_AUTH_USERNAME` | Alias for SASL username. |
| `DEFAULT_CLUSTER_AUTH_PASSWORD` / `KAFKA_MAP_BOOTSTRAP_AUTH_PASSWORD` | Alias for SASL password. |

Example of injecting a default admin and a bootstrap cluster:

```bash
export DEFAULT_USERNAME=ops
export DEFAULT_PASSWORD=supersecret
export DEFAULT_CLUSTER_NAME=prod-kafka
export DEFAULT_CLUSTER_SERVERS="kafka-1:9092,kafka-2:9092"
export DEFAULT_CLUSTER_SECURITY_PROTOCOL=PLAINTEXT
```

Each cluster entry accepts `name`, `servers`, `securityProtocol`, `saslMechanism`, `saslUsername`, and `saslPassword` (or the aliases `authUsername`/`authPassword`). During startup the server validates the connection info and inserts any missing clusters into SQLite, so you can fully provision environments without manual UI steps.

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

### Code Formatting

```bash
make fmt
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
# Build optimized binary (with UPX compression)
make build

# The binary will be created as 'kafka-map-go'
# Deploy with config directory and run
./kafka-map-go
```

## Troubleshooting

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

- Project repository: [Kafka Map Go](https://github.com/bingfengfeifei/kafka-map-go)
- Original Java version: [Kafka Map](https://github.com/dushixiang/kafka-map)
- Built with [Gin](https://github.com/gin-gonic/gin)
- Kafka client: [Sarama](https://github.com/IBM/sarama)
- ORM: [GORM](https://gorm.io)
