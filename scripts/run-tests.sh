#!/bin/bash

# AutoDataHub-monitor 集成测试脚本

set -e

echo "🚀 开始运行 AutoDataHub-monitor 集成测试..."

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker 未运行，请启动 Docker 后重试"
    exit 1
fi

# 停止并清理现有的测试容器
echo "🧹 清理现有的测试环境..."
docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true

# 构建并启动测试环境
echo "🐳 启动测试数据库服务..."
docker-compose -f docker-compose.test.yml up -d mysql-test redis-test

# 等待服务启动
echo "⏳ 等待数据库服务启动..."
for i in {1..30}; do
    if docker-compose -f docker-compose.test.yml exec -T mysql-test mysqladmin ping -h localhost --silent; then
        echo "✅ MySQL 已就绪"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ MySQL 启动超时"
        docker-compose -f docker-compose.test.yml logs mysql-test
        exit 1
    fi
    sleep 2
done

for i in {1..15}; do
    if docker-compose -f docker-compose.test.yml exec -T redis-test redis-cli ping >/dev/null 2>&1; then
        echo "✅ Redis 已就绪"
        break
    fi
    if [ $i -eq 15 ]; then
        echo "❌ Redis 启动超时"
        docker-compose -f docker-compose.test.yml logs redis-test
        exit 1
    fi
    sleep 2
done

# 运行单元测试
echo "🧪 运行单元测试..."
go test -v -race -coverprofile=coverage.out ./pkg/...

# 运行集成测试
echo "🔗 运行集成测试..."
MYSQL_HOST=localhost \
MYSQL_PORT=3307 \
MYSQL_USER=testuser \
MYSQL_PASSWORD=testpassword \
MYSQL_DATABASE=autodatahub \
REDIS_ADDR=localhost:6380 \
go test -v ./tests/integration/...

# 运行基准测试
echo "⚡ 运行基准测试..."
MYSQL_HOST=localhost \
MYSQL_PORT=3307 \
MYSQL_USER=testuser \
MYSQL_PASSWORD=testpassword \
MYSQL_DATABASE=autodatahub \
REDIS_ADDR=localhost:6380 \
go test -bench=. -run=^$ ./tests/integration/...

# 生成测试覆盖率报告
echo "📊 生成测试覆盖率报告..."
go tool cover -html=coverage.out -o coverage.html
echo "📈 测试覆盖率报告已生成: coverage.html"

# 清理测试环境
echo "🧹 清理测试环境..."
docker-compose -f docker-compose.test.yml down -v

echo "✅ 所有测试完成！"
