# AutoDataHub-Monitor

## 项目概述

AutoDataHub-Monitor是一个专门为车辆数据监控设计的高性能实时监控系统。该系统能够从多个数据源（如Kafka、Redis等）收集车辆数据，通过可配置的处理流水线进行分析和转换，并根据预设规则触发警报。系统支持灵活的插件扩展，可以根据不同的业务需求进行定制化开发。

## 系统架构

系统由以下主要组件构成：

```
+----------------+     +-------------+     +--------------+
|   数据源模块    | --> |  处理流水线  | --> |   日志系统    |
+----------------+     +-------------+     +--------------+
        |                     |                  |
        v                     v                  v
+----------------+     +--------------+    +--------------+
|   插件系统     |     |   告警模块   |    |  Redis存储   |
+----------------+     +--------------+    +--------------+
```

### 核心模块

1. **数据源模块**
   - Kafka连接器：从Kafka主题订阅车辆数据
   - Redis连接器：从Redis队列获取数据
   - 触发器API：接收外部系统的数据推送

2. **处理流水线**
   - 车辆类型过滤器：根据车辆类型和使用类型进行数据分流
   - 数据转换器：标准化数据格式
   - 告警处理器：根据业务规则生成告警

3. **插件系统**
   - 插件注册表：管理和加载自定义插件
   - 扩展点：支持数据处理、告警规则等扩展

4. **存储系统**
   - Redis队列：存储处理后的数据
   - 日志存储：记录系统运行和数据处理日志

## 配置指南

系统配置文件位于`configs/`目录下，支持YAML格式：

### Redis配置

```yaml
redis:
  host: "localhost"
  port: 6379
  db: 0
  password: ""
  pool_size: 10
```

### 车辆类型配置

```yaml
vehicle_type:
  default_queue: "default_queue"
  type_map:
    "passenger_private": "passenger_queue"
    "commercial_logistics": "logistics_queue"
    "commercial_passenger": "bus_queue"
```

### 触发器配置

```yaml
trigger:
  api:
    port: 8080
    endpoint: "/api/trigger"
  rules:
    - type: "negative"
      conditions:
        - field: "speed"
          operator: ">"
          value: 120
```

## 使用指南

### 环境要求

- Go 1.23+
- Redis 6.0+
- Kafka 2.8+

### 安装

```bash
# 克隆项目
git clone https://github.com/zhangyuchen/AutoDataHub-monitor.git

# 安装依赖
go mod download

# 编译
go build -o bin/monitor cmd/monitor/main.go
```

### 启动服务

```bash
# 使用默认配置启动
./bin/monitor

# 指定配置文件启动
./bin/monitor --config=configs/custom.yaml
```

### API使用示例

```bash
# 发送触发器数据
curl -X POST http://localhost:8080/api/trigger \
  -H "Content-Type: application/json" \
  -d '{
    "vin": "VIN12345",
    "car_type": "passenger",
    "usage_type": "private",
    "trigger_id": "speed_alert",
    "type": "negative",
    "timestamp": 1623456789
  }'
```

## 开发指南

### 添加新的处理器

1. 在`internal/processor`目录下创建新的处理器包
2. 实现处理器接口
3. 在处理器注册表中注册新处理器

```go
package myprocessor

import (
    "github.com/zhangyuchen/AutoDataHub-monitor/pkg/models"
)

type MyProcessor struct {
    // 处理器配置
}

func (p *MyProcessor) Process(data *models.NegativeTriggerData) error {
    // 实现数据处理逻辑
    return nil
}
```

### 添加新的数据源

1. 在`internal/datasource`目录下创建新的数据源包
2. 实现数据源接口
3. 在数据源注册表中注册新数据源

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交Pull Request

## 许可证

本项目采用MIT许可证。详情请参阅LICENSE文件。