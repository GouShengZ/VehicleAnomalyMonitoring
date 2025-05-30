#!/bin/bash

# AutoDataHub-monitor 本地开发环境启动脚本

set -e

echo "🚀 启动 AutoDataHub-monitor 开发环境..."

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker 未运行，请启动 Docker 后重试"
    exit 1
fi

# 启动开发环境数据库
echo "🐳 启动开发环境数据库..."
docker-compose up -d mysql redis

# 等待服务启动
echo "⏳ 等待数据库服务启动..."
for i in {1..30}; do
    if docker-compose exec mysql mysqladmin ping -h localhost --silent 2>/dev/null; then
        echo "✅ MySQL 已就绪"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ MySQL 启动超时"
        docker-compose logs mysql
        exit 1
    fi
    sleep 2
done

for i in {1..15}; do
    if docker-compose exec redis redis-cli ping >/dev/null 2>&1; then
        echo "✅ Redis 已就绪"
        break
    fi
    if [ $i -eq 15 ]; then
        echo "❌ Redis 启动超时"
        docker-compose logs redis
        exit 1
    fi
    sleep 2
done

# 运行数据库迁移（如果有的话）
echo "🗄️ 初始化数据库..."
# 这里可以添加数据库迁移命令

# 启动应用
echo "🌟 启动应用..."
go run ./cmd/main.go
