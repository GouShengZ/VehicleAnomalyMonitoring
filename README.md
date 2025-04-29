# AutoDataHub-monitor

## 项目简介

AutoDataHub-monitor 是一个用于车辆异常监控的分布式系统。该系统通过消息队列实现各个处理节点之间的解耦，支持水平扩展，主要用于监控和处理车辆运行过程中的异常情况。
该项目整体逻辑经过量产车产测试，稳定性良好。

## 系统架构

### 整体架构

系统采用分布式消息队列架构，主要包含以下组件：

- **数据源节点**：负责从上游系统获取车辆异常触发数据
- **处理节点**：执行具体的数据处理和分析逻辑
- **告警节点**：监控队列状态并发送告警信息
- **消息队列**：默认使用Redis作为消息队列，支持扩展其他队列组件

### 目录结构

```
├── cmd                     # 启动入口
│   ├── process            # 处理服务入口
│   └── task               # 定时任务入口
├── configs                # 配置文件
├── internal               # 内部实现
│   ├── datasource         # 上游数据源适配
│   ├── pipeline           # 节点流水线
│   └── processor          # 处理器实现
├── pkg                    # 公共包
│   ├── models             # 数据模型
│   └── utils              # 工具函数
└── scripts                # 脚本文件
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

## 启动方式

### 处理服务

```bash
# 启动处理服务
go run cmd/process/process_main.go
```

### 定时任务

```bash
# 启动定时任务
go run cmd/task/task_main.go
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

- Redis: 消息队列和数据缓存
- MySQL: 数据持久化存储
- 飞书: 告警通知

## 注意事项

1. 确保Redis和MySQL服务正常运行
2. 配置文件中的敏感信息需要proper管理
3. 建议对处理节点进行监控
4. 定期检查日志文件大小