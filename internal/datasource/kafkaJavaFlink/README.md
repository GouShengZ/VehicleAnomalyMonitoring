# Vehicle Anomaly Detection - Java Flink Project

这是一个基于Apache Flink的车辆异常检测系统，用于消费Kafka数据流，进行实时异常检测，并将结果存储到Redis队列或MySQL数据库。

## 功能特性

- **Kafka消费**: 从Kafka主题消费车辆数据
- **实时处理**: 使用Apache Flink进行实时数据处理
- **异常检测**: 检测车辆数据中的各种异常情况
- **双重存储**: 优先写入Redis队列，失败时写入MySQL数据库
- **容错机制**: 支持检查点和故障恢复

## 异常检测规则

系统检测以下异常类型：

1. **速度异常检测**:
   - **1秒内速度变化** (SPEED_CHANGE_1S): 速度变化超过18 m/s
   - **2秒内速度变化** (SPEED_CHANGE_2S): 速度变化超过24 m/s

2. **系统信号异常**:
   - **AEB触发** (AEB_TRIGGERED): 自动紧急制动系统激活
   - **NOA退出** (NOA_EXIT): 导航辅助驾驶系统退出
   - **ACC退出** (ACC_EXIT): 自适应巡航控制系统退出
   - **LKA退出** (LKA_EXIT): 车道保持辅助系统退出

## 数据模型

### 车辆数据模型 (VehicleData)
- `vin`: 车辆VIN码
- `kafkaTimestamp`: Kafka时间戳（进入Kafka的时间）
- `sigTimestamp`: 信号时间戳（信号触发时间）
- `sig`: 信号数据对象，包含：
  - `sigName`: 信号名称（如speed、AEB、NOA_EXIT等）
  - `value`: 信号值（类型可变）

### 异常数据模型 (AnomalyData)
- `vin`: 车辆VIN码
- `kafkaTimestamp`: Kafka时间戳
- `sigTimestamp`: 信号时间戳
- `anomalyType`: 异常类型
- `severity`: 严重程度（HIGH、MEDIUM、LOW）
- `description`: 异常描述
- `sigName`: 触发异常的信号名称
- `currentValue`: 当前值
- `thresholdValue`: 阈值
- `additionalInfo`: 附加信息

## 项目结构

```
├── src/main/java/com/vehicle/anomaly/
│   ├── FlinkKafkaProcessor.java          # 主程序入口
│   ├── config/
│   │   └── ConfigManager.java            # 配置管理
│   ├── model/
│   │   ├── VehicleData.java             # 车辆数据模型
│   │   └── AnomalyData.java             # 异常数据模型
│   ├── deserializer/
│   │   └── VehicleDataDeserializer.java # Kafka消息反序列化
│   ├── processor/
│   │   └── AnomalyDetectionProcessor.java # 异常检测处理器
│   ├── sink/
│   │   └── AnomalyDataSink.java         # 数据输出Sink
│   └── util/
│       └── TestDataGenerator.java       # 测试数据生成器
├── src/main/resources/
│   ├── application.properties           # 配置文件
│   └── logback.xml                     # 日志配置
├── sql/
│   └── init.sql                        # 数据库初始化脚本
├── pom.xml                             # Maven项目配置
└── start.sh                           # 启动脚本
```

## 环境要求

- Java 11 或更高版本
- Maven 3.6 或更高版本
- Apache Kafka 2.8+
- Redis 5.0+
- MySQL 8.0+

## 配置说明

编辑 `src/main/resources/application.properties` 文件：

```properties
# Kafka配置
kafka.bootstrap.servers=localhost:9092
kafka.group.id=vehicle-anomaly-group
kafka.topic=vehicle-data

# Redis配置
redis.host=localhost
redis.port=6379
redis.password=
redis.database=0
redis.timeout=5000
redis.queue.key=vehicle_anomaly_queue

# MySQL配置
mysql.host=localhost
mysql.port=3306
mysql.database=vehicle_anomaly
mysql.username=root
mysql.password=password
mysql.table=anomaly_data
```

## 安装和运行

### 1. 准备环境

确保Kafka、Redis和MySQL服务正在运行。

### 2. 初始化数据库

```bash
mysql -u root -p < sql/init.sql
```

### 3. 编译项目

```bash
mvn clean package -DskipTests
```

### 4. 运行应用

```bash
# 使用启动脚本
chmod +x start.sh
./start.sh

# 或者直接运行
java -cp target/kafka-flink-processor-1.0.0.jar com.vehicle.anomaly.FlinkKafkaProcessor
```

## 测试

### 生成测试数据

运行测试数据生成器向Kafka发送模拟车辆数据：

```bash
java -cp target/kafka-flink-processor-1.0.0.jar com.vehicle.anomaly.util.TestDataGenerator
```

### 查看结果

1. **Redis队列**: 检查Redis中的异常数据
```bash
redis-cli
> LRANGE vehicle_anomaly_queue 0 -1
```

2. **MySQL数据库**: 查询异常数据表
```sql
SELECT * FROM anomaly_data ORDER BY timestamp DESC;
```

## 数据格式

### 输入数据格式 (Kafka消息)

```json
{
  "vin": "1HGCM82633A123456",
  "kafkaTimestamp": 1704067200000,
  "sigTimestamp": 1704067199500,
  "sig": {
    "sigName": "speed",
    "value": 25.5
  }
}
```

### 输出数据格式 (异常数据)

```json
{
  "vin": "1HGCM82633A123456",
  "kafkaTimestamp": 1704067200000,
  "sigTimestamp": 1704067199500,
  "anomalyType": "SPEED_CHANGE_1S",
  "severity": "HIGH",
  "description": "Speed changed too rapidly within 1 second",
  "sigName": "speed",
  "currentValue": 25.5,
  "thresholdValue": 18.0,
  "additionalInfo": "Previous speed: 5.2 m/s, Current speed: 25.5 m/s, Time diff: 800 ms"
}
```

## 监控和日志

- 日志文件位于 `logs/vehicle-anomaly-detection.log`
- 可通过修改 `logback.xml` 调整日志级别
- 支持Flink Web UI监控 (默认端口8081)

## 扩展功能

1. **自定义异常规则**: 修改 `AnomalyDetectionProcessor` 类
2. **添加新的数据源**: 实现新的Source连接器
3. **增加告警功能**: 在Sink中添加告警逻辑
4. **优化性能**: 调整并行度和资源配置

## 故障排除

1. **Kafka连接问题**: 检查Kafka服务状态和配置
2. **Redis连接问题**: 确认Redis服务可用，检查连接参数
3. **MySQL连接问题**: 验证数据库连接信息和权限
4. **内存不足**: 调整JVM参数或Flink配置

## 许可证

本项目采用Apache License 2.0许可证。
