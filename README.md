# AutoDataHub-monitor

> 该项目是博主复刻自己曾经在工作中的项目，通过AI完成，只能借鉴大概的想法，并不适用于所有环境

## 项目简介

AutoDataHub-monitor 是一个用于车辆异常监控的分布式系统。该系统通过消息队列实现各个处理节点之间的解耦，支持水平扩展，主要用于监控和处理车辆运行过程中的异常情况。
该项目整体逻辑经过量产车产测试，稳定性良好。

## 特性

- 🚀 **高性能**: 基于 Go 语言开发，支持高并发处理
- 🔧 **可扩展**: 模块化设计，支持水平扩展
- 📊 **可观测**: 完整的监控、日志和健康检查
- 🔄 **高可用**: 完善的重试机制和错误处理
- 🐳 **容器化**: 支持 Docker 部署
- 🧪 **测试覆盖**: 完整的单元测试和集成测试

## 快速开始

### 环境要求

- Go 1.21+
- Docker & Docker Compose
- MySQL 8.0+
- Redis 7.0+

### 安装和运行

1. **克隆项目**
```bash
git clone <repository-url>
cd AutoDataHub-monitor
```

2. **配置环境**
```bash
# 复制环境配置文件
cp .env.example .env

# 编辑配置文件（根据实际环境修改）
vim .env
```

3. **启动开发环境**
```bash
# 使用 Makefile（推荐）
make dev-setup

# 或手动启动
./scripts/dev-setup.sh
```

4. **运行测试**
```bash
# 运行所有测试
make test-all

# 只运行单元测试
make test-unit

# 运行集成测试
make test-integration
```

### Docker 部署

```bash
# 构建和运行
docker-compose up -d

# 运行测试
make docker-test
```

## 项目架构

### 整体架构

系统采用分布式消息队列架构，主要包含以下组件：

- **数据源节点**：负责从上游系统获取车辆异常触发数据
- **处理节点**：执行具体的数据处理和分析逻辑
- **告警节点**：监控队列状态并发送告警信息
- **消息队列**：默认使用Redis作为消息队列，支持扩展其他队列组件

### 目录结构

```
├── cmd/                    # 启动入口
│   ├── main.go            # 主程序入口
│   ├── process/           # 处理服务入口（旧版本）
│   └── task/              # 定时任务入口（旧版本）
├── configs/               # 配置文件
├── internal/              # 内部实现
│   ├── datasource/        # 上游数据源适配
│   ├── pipeline/          # 节点流水线
│   └── processor/         # 处理器实现
├── pkg/                   # 公共包
│   ├── health/            # 健康检查
│   ├── metrics/           # 监控指标
│   ├── models/            # 数据模型
│   └── utils/             # 工具函数
├── tests/                 # 测试文件
│   └── integration/       # 集成测试
├── scripts/               # 脚本文件
├── .github/workflows/     # CI/CD 配置
└── docker-compose.*.yml   # Docker 配置
```

## 核心功能

### 1. 数据源接入 (internal/datasource)
- 支持多种数据源接入
- 实现了触发器数据源适配
- 支持车辆信息数据源接入

### 2. 数据处理节点 (internal/processor/node)
- 独立的处理单元
- 支持自定义处理逻辑
- 处理结果可传递至下游节点

### 3. 处理流水线 (internal/pipeline)
- 组合多个处理节点
- 支持并行处理
- 灵活的节点编排

### 4. 监控告警 (internal/processor/alert)
- 队列长度监控
- 飞书告警集成
- 可配置告警阈值

### 5. 监控指标 (pkg/metrics)
- Prometheus指标收集
- 消息处理、队列状态、系统资源监控
- HTTP指标服务器 (:8080/metrics)
- 数据库和Redis连接监控

### 6. 健康检查 (pkg/health)
- 系统健康状态监控
- 依赖服务状态检查
- HTTP健康检查接口 (:8080/health)

### 7. 重试机制 (pkg/utils)
- 指数退避重试
- 可配置重试策略
- 非重试错误支持

### 8. 配置管理 (configs)
- 环境变量支持
- 默认配置机制
- 配置文件热加载
- 敏感信息环境变量配置

## 数据流转

系统核心数据结构为 `NegativeTriggerData`，包含以下关键信息：

```go
type NegativeTriggerData struct {
    Vin          string // 车辆识别号
    Timestamp    int64  // 触发时间戳
    CarType      string // 车辆类型
    UsageType    string // 使用类型
    TriggerID    string // 触发器ID
    LogId        int    // 日志ID
    ThresholdLog string // 阈值日志
    IsCrash      int    // 是否发生碰撞
}
```

## 配置说明

主要配置文件位于 `configs` 目录：

- `config.yaml`: 系统主配置文件
- `can_sig_*.yaml`: 不同场景的CAN信号配置  
- `steering_angle.dbc`: 转向角度解析配置
- `.env`: 环境变量配置文件

### 环境变量

复制 `.env.example` 为 `.env` 并根据实际环境修改：

```bash
# 数据库配置
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=autodatahub
MYSQL_PASSWORD=your_password
MYSQL_DATABASE=autodatahub

# Redis 配置
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# 服务配置
SERVER_PORT=8080
HEALTH_CHECK_PORT=8081
LOG_LEVEL=info
```

## 启动方式

### 开发环境

#### 处理服务

```bash
# 启动处理服务
go run cmd/process/process_main.go
```

#### 定时任务

```bash
# 启动定时任务
go run cmd/task/task_main.go
```

### 生产环境

#### 使用Docker Compose启动

```bash
# 启动所有服务（Redis、MySQL、处理服务、定时任务）
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f autodatahub-process
docker-compose logs -f autodatahub-task
```

#### 环境变量配置

```bash
# 复制环境变量配置文件
cp .env.example .env

# 编辑环境变量
vim .env
```

支持的环境变量：
- `REDIS_ADDR`: Redis地址 (默认: localhost:6379)
- `REDIS_PASSWORD`: Redis密码
- `MYSQL_HOST`: MySQL主机 (默认: localhost)
- `MYSQL_PORT`: MySQL端口 (默认: 3306)
- `MYSQL_USER`: MySQL用户名 (默认: root)
- `MYSQL_PASSWORD`: MySQL密码
- `MYSQL_DATABASE`: MySQL数据库名 (默认: autodatahub)
- `LOG_LEVEL`: 日志级别 (默认: info)
- `QUEUE_SIZE`: 队列大小 (默认: 1000)

### 监控和健康检查

```bash
# 健康检查
curl http://localhost:8080/health

# Prometheus指标
curl http://localhost:8080/metrics
```

## 扩展性设计

1. **节点解耦**
   - 各处理节点通过消息队列通信
   - 支持节点的独立部署和扩展
   - 可以根据负载动态调整节点数量

2. **队列适配**
   - 默认使用Redis作为消息队列
   - 支持扩展其他队列组件
   - 统一的队列接口设计

3. **处理器扩展**
   - 支持自定义处理节点
   - 灵活的节点组合方式
   - 标准化的数据输入输出

## 依赖组件

- **Redis**: 消息队列和数据缓存
- **MySQL**: 数据持久化存储
- **飞书**: 告警通知
- **Prometheus**: 监控指标收集 (可选)

## 监控指标

系统提供丰富的监控指标，可通过 `/metrics` 端点获取：

### 消息处理指标
- `message_processed_total`: 已处理消息总数
- `message_processing_duration_seconds`: 消息处理耗时

### 队列监控指标
- `queue_length`: 队列长度
- `queue_produced_total`: 队列生产消息总数
- `queue_consumed_total`: 队列消费消息总数

### 系统资源指标
- `system_cpu_usage_percent`: CPU使用率
- `system_memory_usage_bytes`: 内存使用量

### 数据库指标
- `database_connections_active`: 活跃数据库连接数
- `database_query_duration_seconds`: 数据库查询耗时

### Redis指标
- `redis_connections_active`: 活跃Redis连接数
- `redis_operation_duration_seconds`: Redis操作耗时

## 测试

```bash
# 运行所有测试
go test ./...

# 运行特定测试文件
go test ./pkg/utils/retry_test.go
go test ./pkg/health/health_test.go

# 运行测试并查看覆盖率
go test -cover ./...

# 运行基准测试
go test -bench=. ./pkg/utils/
```

## 注意事项

1. **服务依赖**
   - 确保Redis和MySQL服务正常运行
   - 建议使用Docker Compose统一管理依赖服务

2. **配置管理**
   - 配置文件中的敏感信息建议使用环境变量
   - 生产环境建议使用配置管理工具

3. **监控运维**
   - 建议对处理节点进行监控
   - 定期检查日志文件大小
   - 监控队列长度和处理延迟

4. **性能优化**
   - 根据业务量调整队列大小
   - 监控CPU和内存使用情况
   - 适当调整并发处理数量

5. **故障排查**
   - 查看 `/health` 接口确认服务状态
   - 通过 `/metrics` 接口监控关键指标
   - 检查应用日志和依赖服务状态

## API文档

### 健康检查接口

```http
GET /health
```

返回示例：
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "checks": {
    "redis": "healthy",
    "mysql": "healthy"
  }
}
```

### 监控指标接口

```http
GET /metrics
```

返回Prometheus格式的监控指标。