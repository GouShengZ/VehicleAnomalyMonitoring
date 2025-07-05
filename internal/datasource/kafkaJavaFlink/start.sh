#!/bin/bash

# 启动脚本

echo "Starting Vehicle Anomaly Detection System..."

# 检查Java环境
if ! command -v java &> /dev/null; then
    echo "Java is not installed. Please install Java 11 or higher."
    exit 1
fi

# 检查Maven环境
if ! command -v mvn &> /dev/null; then
    echo "Maven is not installed. Please install Maven 3.6 or higher."
    exit 1
fi

# 编译项目
echo "Building project..."
mvn clean package -DskipTests

if [ $? -ne 0 ]; then
    echo "Build failed. Please check the errors above."
    exit 1
fi

# 运行应用
echo "Starting Flink job..."
java -cp target/kafka-flink-processor-1.0.0.jar com.vehicle.anomaly.FlinkKafkaProcessor

echo "Application started successfully!"
