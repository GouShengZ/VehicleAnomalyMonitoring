# AutoDataHub-monitor

车辆状态监控系统

## 功能特点

1. 支持设置不同车辆类型
2. 可配置监控阈值
3. 多种判断函数（CAN信号判断、感知数据判断等）
4. 双数据源支持：
   - 秒级CAN信号（Kafka）
   - 车辆采集的trigger数据
5. 支持通过DBC文件解析CAN文件获取毫秒级信号数据

## 使用方法

### 在线模式（实时监控）

```bash
go run cmd/monitor/main.go --config=configs/config.yaml
```

### 离线模式（解析CAN文件）

```bash
go run cmd/monitor/main.go --config=configs/config.yaml --can-file=/path/to/can/file.can --vehicle-id=car001 --vehicle-type=sedan
```

## 配置说明

在`configs/config.yaml`中可以配置：

1. Kafka连接信息
2. 车辆类型定义
3. 监控阈值设置
4. 数据处理参数
5. 告警配置
6. DBC文件路径

## 项目结构

```
/
├── cmd/                    # 应用程序入口
│   └── monitor/            # 监控服务入口
├── configs/                # 配置文件
├── internal/               # 内部包
│   ├── models/             # 数据模型
│   ├── parser/             # 数据解析器
│   │   ├── can/            # CAN数据解析
│   │   └── dbc/            # DBC文件解析
│   ├── monitor/            # 监控逻辑
│   └── consumer/           # 数据消费者
├── pkg/                    # 公共包
│   ├── kafka/              # Kafka客户端
│   └── utils/              # 工具函数
└── test/                   # 测试文件
```