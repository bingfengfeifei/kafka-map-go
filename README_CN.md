# Kafka Map Go

[English](README.md) | 简体中文

一个用 Go 语言编写的 Kafka 可视化和管理工具，从原始的 Java Spring Boot 版本转换而来。

## 功能特性

- **多集群管理** - 添加、删除和管理多个 Kafka 集群
- **集群监控** - 查看集群状态、分区数、副本数、存储大小、偏移量
- **主题管理** - 创建、删除、扩展分区、查看配置
- **Broker 监控** - 查看 Broker 状态和配置
- **消费者组管理** - 查看、删除消费者组、监控延迟
- **偏移量管理** - 重置消费者组偏移量（开始、结束或指定偏移量）
- **消息操作** - 查看、搜索和发送消息到主题
- **身份认证** - 基于令牌的身份认证和密码管理
- **内嵌前端** - React 前端内嵌在单个二进制文件中

## 技术栈

### 后端
- **Gin** - Web 框架
- **GORM** - 数据库 ORM
- **Sarama** - Kafka 客户端库
- **SQLite** - 内嵌数据库（无 CGO 依赖）
- **BCrypt** - 密码加密

### 前端
- **React** - UI 框架
- **Vite** - 构建工具
- 使用 Go 的 `embed` 包内嵌

## 环境要求

- Go 1.21 或更高版本
- Node.js 18+ 和 npm（用于构建前端）

## 快速开始

### 使用 Make

```bash
# 安装依赖并构建
make install-deps
make build

# 运行应用程序
make run

# 或以开发模式运行
make dev
```

### 手动构建

```bash
# 安装 Go 依赖
go mod download

# 构建前端
cd web
npm install --legacy-peer-deps
npm run build
cd ..

# 复制前端资源
mkdir -p cmd/server/web
cp -r web/index.html web/assets cmd/server/web/

# 构建 Go 应用程序
go build -o kafka-map-go ./cmd/server

# 运行
./kafka-map-go
```

### 使用 Docker

```bash
# 构建 Docker 镜像
docker build -t kafka-map-go:latest .

# 运行容器
docker run -p 8080:8080 -v $(pwd)/data:/app/data kafka-map-go:latest
```

## 配置

编辑 `config/config.yaml`：

```yaml
server:
  port: 8080

database:
  path: data/kafka-map.db

default:
  username: admin
  password: admin

cache:
  token_expiration: 7200  # 2 小时（秒）
  max_tokens: 100
```

### 环境变量

您可以通过以下环境变量覆盖任何 YAML 设置（或提供新的默认值）：

| 变量 | 描述 |
| --- | --- |
| `KAFKA_MAP_SERVER_PORT` | 覆盖 `server.port`。 |
| `KAFKA_MAP_DATABASE_PATH` | 覆盖 `database.path`。 |
| `DEFAULT_USERNAME` / `KAFKA_MAP_DEFAULT_USERNAME` | 覆盖初始管理员用户名。 |
| `DEFAULT_PASSWORD` / `KAFKA_MAP_DEFAULT_PASSWORD` | 覆盖初始管理员密码。 |
| `KAFKA_MAP_CACHE_TOKEN_EXPIRATION` | 覆盖 `cache.token_expiration`（秒）。 |
| `KAFKA_MAP_CACHE_MAX_TOKENS` | 覆盖 `cache.max_tokens`。 |
| `DEFAULT_CLUSTER_NAME` / `KAFKA_MAP_BOOTSTRAP_NAME` | 启动时自动创建的集群名称。 |
| `DEFAULT_CLUSTER_SERVERS` / `KAFKA_MAP_BOOTSTRAP_SERVERS` | 引导集群的逗号分隔的 Broker 列表。 |
| `DEFAULT_CLUSTER_SECURITY_PROTOCOL` / `KAFKA_MAP_BOOTSTRAP_SECURITY_PROTOCOL` | 可选的安全协议（默认为 `PLAINTEXT`）。 |
| `DEFAULT_CLUSTER_SASL_MECHANISM` / `KAFKA_MAP_BOOTSTRAP_SASL_MECHANISM` | 可选的 SASL 机制。 |
| `DEFAULT_CLUSTER_SASL_USERNAME` / `KAFKA_MAP_BOOTSTRAP_SASL_USERNAME` | SASL 用户名（如果需要）。 |
| `DEFAULT_CLUSTER_SASL_PASSWORD` / `KAFKA_MAP_BOOTSTRAP_SASL_PASSWORD` | SASL 密码（如果需要）。 |
| `DEFAULT_CLUSTER_AUTH_USERNAME` / `KAFKA_MAP_BOOTSTRAP_AUTH_USERNAME` | SASL 用户名的别名。 |
| `DEFAULT_CLUSTER_AUTH_PASSWORD` / `KAFKA_MAP_BOOTSTRAP_AUTH_PASSWORD` | SASL 密码的别名。 |

注入默认管理员和引导集群的示例：

```bash
export DEFAULT_USERNAME=ops
export DEFAULT_PASSWORD=supersecret
export DEFAULT_CLUSTER_NAME=prod-kafka
export DEFAULT_CLUSTER_SERVERS="kafka-1:9092,kafka-2:9092"
export DEFAULT_CLUSTER_SECURITY_PROTOCOL=PLAINTEXT
```

每个集群条目接受 `name`、`servers`、`securityProtocol`、`saslMechanism`、`saslUsername` 和 `saslPassword`（或别名 `authUsername`/`authPassword`）。在启动期间，服务器会验证连接信息并将任何缺失的集群插入 SQLite，因此您可以完全配置环境而无需手动 UI 步骤。

## 默认凭据

- **用户名**: admin
- **密码**: admin

**重要提示**：首次登录后请更改默认密码！

## API 端点

### 身份认证
- `POST /api/login` - 用户登录
- `POST /api/logout` - 用户登出
- `GET /api/info` - 获取当前用户信息
- `POST /api/change-password` - 修改密码

### 集群
- `GET /api/clusters` - 列出所有集群
- `GET /api/clusters/:id` - 获取集群详情
- `POST /api/clusters` - 创建集群
- `PUT /api/clusters/:id` - 更新集群
- `DELETE /api/clusters/:id` - 删除集群

### Brokers
- `GET /api/brokers?clusterId=:id` - 列出 Brokers
- `GET /api/brokers/:id/configs?clusterId=:id` - 获取 Broker 配置
- `PUT /api/brokers/:id/configs?clusterId=:id` - 更新 Broker 配置

### 主题
- `GET /api/topics?clusterId=:id` - 列出主题
- `GET /api/topics/names?clusterId=:id` - 获取主题名称
- `GET /api/topics/:topic?clusterId=:id` - 获取主题详情
- `POST /api/topics?clusterId=:id` - 创建主题
- `POST /api/topics/batch-delete?clusterId=:id` - 删除主题
- `POST /api/topics/:topic/partitions?clusterId=:id` - 扩展分区
- `PUT /api/topics/:topic/configs?clusterId=:id` - 更新主题配置
- `GET /api/topics/:topic/data?clusterId=:id` - 获取消息
- `POST /api/topics/:topic/data?clusterId=:id` - 发送消息

### 消费者组
- `GET /api/consumerGroups?clusterId=:id` - 列出消费者组
- `GET /api/consumerGroups/:groupId?clusterId=:id` - 获取消费者组详情
- `DELETE /api/consumerGroups/:groupId?clusterId=:id` - 删除消费者组
- `GET /api/topics/:topic/consumerGroups?clusterId=:id` - 按主题获取消费者组
- `GET /api/topics/:topic/consumerGroups/:groupId/offset?clusterId=:id` - 获取偏移量
- `PUT /api/topics/:topic/consumerGroups/:groupId/offset?clusterId=:id` - 重置偏移量

## 项目结构

```
kafka-map-go/
├── cmd/
│   └── server/
│       ├── main.go              # 应用程序入口
│       └── web/                 # 内嵌的前端资源
├── internal/
│   ├── config/                  # 配置管理
│   ├── controller/              # HTTP 处理器
│   ├── dto/                     # 数据传输对象
│   ├── middleware/              # 中间件（认证、CORS）
│   ├── model/                   # 数据库模型
│   ├── repository/              # 数据访问层
│   ├── service/                 # 业务逻辑
│   └── util/                    # 工具类
├── pkg/
│   └── database/                # 数据库初始化
├── web/                         # 前端源代码
├── config/
│   └── config.yaml              # 配置文件
├── Dockerfile
├── Makefile
└── README.md
```

## 开发

### 运行测试

```bash
make test
```

### 代码格式化

```bash
make fmt
```

### 代码检查

```bash
make lint
```

### 清理构建产物

```bash
make clean
```

## 安全特性

- **基于令牌的身份认证** - 2 小时令牌过期时间
- **BCrypt 密码哈希** - 安全的密码存储
- **CORS 支持** - 可配置的跨域请求
- **SASL/SCRAM 支持** - 安全的 Kafka 身份认证（PLAIN、SCRAM-SHA-256、SCRAM-SHA-512）
- **SSL/TLS 支持** - 加密的 Kafka 连接

## 支持的 Kafka 安全协议

- PLAINTEXT
- SASL_PLAINTEXT
- SASL_SSL
- SSL

## 支持的 SASL 机制

- PLAIN
- SCRAM-SHA-256
- SCRAM-SHA-512

## 与 Java 版本的差异

### 未包含的功能
- **延迟消息系统** - 此版本未实现 18 级延迟消息功能

### 改进之处
- **单一二进制文件** - 前端内嵌在 Go 二进制文件中
- **更小的占用空间** - 无需 JVM
- **更快的启动速度** - 原生二进制执行
- **更简单的部署** - 单个可执行文件

## 生产环境构建

```bash
# 构建优化的二进制文件
make build-prod

# 二进制文件将创建为 'kafka-map-go'
# 与 config 目录一起部署并运行
./kafka-map-go
```

## 故障排除

### 前端构建问题
如果前端构建失败，请尝试：
```bash
cd web
rm -rf node_modules package-lock.json
npm install --legacy-peer-deps
npm run build
```

## 贡献

欢迎贡献！请随时提交 Pull Request。

## 许可证

本项目从原始 Java 版本的 Kafka-Map 转换而来。

## 致谢

- 项目仓库：[Kafka Map Go](https://github.com/bingfengfeifei/kafka-map-go)
- 原始 Java 版本：[Kafka Map](https://github.com/dushixiang/kafka-map)
- 使用 [Gin](https://github.com/gin-gonic/gin) 构建
- Kafka 客户端：[Sarama](https://github.com/IBM/sarama)
- ORM：[GORM](https://gorm.io)
