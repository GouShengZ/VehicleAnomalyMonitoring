# 数据库初始化脚本

# 创建数据库
CREATE DATABASE IF NOT EXISTS vehicle_anomaly CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 使用数据库
USE vehicle_anomaly;

# 创建异常数据表
CREATE TABLE IF NOT EXISTS anomaly_data (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    vin VARCHAR(17) NOT NULL COMMENT '车辆VIN码',
    kafka_timestamp BIGINT NOT NULL COMMENT 'Kafka时间戳',
    sig_timestamp BIGINT NOT NULL COMMENT '信号时间戳',
    anomaly_type VARCHAR(50) NOT NULL COMMENT '异常类型',
    severity VARCHAR(20) NOT NULL COMMENT '严重程度',
    description TEXT COMMENT '异常描述',
    sig_name VARCHAR(50) NOT NULL COMMENT '信号名称',
    current_value TEXT COMMENT '当前值',
    threshold_value TEXT COMMENT '阈值',
    additional_info TEXT COMMENT '附加信息',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_vin (vin),
    INDEX idx_kafka_timestamp (kafka_timestamp),
    INDEX idx_sig_timestamp (sig_timestamp),
    INDEX idx_anomaly_type (anomaly_type),
    INDEX idx_severity (severity),
    INDEX idx_sig_name (sig_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='车辆异常数据表';
