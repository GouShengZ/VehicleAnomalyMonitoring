# AutoDataHub-monitor Makefile

.PHONY: help build test test-unit test-integration test-all bench clean dev-setup docker-build docker-test lint fmt vet deps coverage run-process run-task docker-run

# 默认目标
.DEFAULT_GOAL := help

# Go 相关变量
GO_VERSION := 1.21
BINARY_NAME := autodatahub-monitor
BUILD_DIR := bin
COVERAGE_FILE := coverage.out
APP_NAME = autodatahub-monitor
VERSION ?= latest
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Docker 相关变量
DOCKER_IMAGE := autodatahub-monitor
DOCKER_TAG := latest

# 帮助信息
help: ## 显示帮助信息
	@echo "AutoDataHub-monitor 开发工具"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 默认目标（保持向后兼容）
all: deps fmt vet test-unit build

# 安装依赖
deps: ## 安装依赖
	@echo "📦 安装依赖..."
	@go mod tidy
	@go mod download
	@echo "✅ 依赖安装完成"

# 格式化代码
fmt: ## 格式化代码
	@echo "🎨 格式化代码..."
	@go fmt ./...
	@echo "✅ 代码格式化完成"

# 代码检查
vet: ## 代码检查
	@echo "🔍 进行代码检查..."
	@go vet ./...
	@echo "✅ 代码检查完成"

# 构建相关
build: ## 构建主应用程序
	@echo "🔨 构建主应用程序..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go
	@echo "✅ 构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

build-old: ## 构建旧版本程序（process 和 task）
	@echo "🔨 构建旧版本程序..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/process cmd/process/process_main.go
	@go build -o $(BUILD_DIR)/task cmd/task/task_main.go
	@echo "✅ 旧版本构建完成"

build-linux: ## 构建 Linux 版本
	@echo "🔨 构建 Linux 版本..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux ./cmd/main.go
	@echo "✅ Linux 版本构建完成: $(BUILD_DIR)/$(BINARY_NAME)-linux"

# 测试相关
test-unit: ## 运行单元测试
	@echo "🧪 运行单元测试..."
	@go test -v -short ./pkg/...
	@echo "✅ 单元测试完成"

test-integration: ## 运行集成测试
	@echo "🧪 运行集成测试..."
	@docker-compose -f docker-compose.test.yml up -d
	@sleep 10
	@go test -v ./tests/integration/...
	@docker-compose -f docker-compose.test.yml down
	@echo "✅ 集成测试完成"

test-all: test-unit test-integration ## 运行所有测试

test: test-unit ## 默认运行单元测试

# 性能测试
bench: ## 运行性能测试
	@echo "⚡ 运行性能测试..."
	@go test -bench=. -benchmem ./...
	@echo "✅ 性能测试完成"

# 覆盖率报告
coverage: ## 生成测试覆盖率报告
	@echo "📊 生成测试覆盖率报告..."
	@go test -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "✅ 覆盖率报告生成完成: coverage.html"

# 代码质量
lint: ## 运行代码检查
	@echo "🔍 运行代码检查..."
	@golangci-lint run ./...
	@echo "✅ 代码检查完成"

# 运行应用
run-process: ## 运行处理服务
	@echo "🚀 启动处理服务..."
	@go run cmd/process/process_main.go

run-task: ## 运行定时任务
	@echo "🚀 启动定时任务..."
	@go run cmd/task/task_main.go

run: ## 运行主应用程序
	@echo "🚀 启动主应用程序..."
	@go run cmd/main.go

# Docker 相关
docker-build: ## 构建 Docker 镜像
	@echo "🐳 构建 Docker 镜像..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Docker 镜像构建完成"

docker-test: ## 运行 Docker 测试环境
	@echo "🐳 启动 Docker 测试环境..."
	@docker-compose -f docker-compose.test.yml up --build

docker-run: ## 运行 Docker 容器
	@echo "🐳 运行 Docker 容器..."
	@docker-compose up -d

docker-stop: ## 停止 Docker 容器
	@echo "🐳 停止 Docker 容器..."
	@docker-compose down

# 开发环境
dev-setup: ## 设置开发环境
	@echo "🔧 设置开发环境..."
	@mkdir -p logs bin
	@go mod tidy
	@go mod download
	@echo "✅ 开发环境设置完成"

install-tools: ## 安装开发工具
	@echo "🔧 安装开发工具..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ 开发工具安装完成"

# 清理
clean: ## 清理构建文件
	@echo "🧹 清理构建文件..."
	@rm -rf $(BUILD_DIR)/
	@rm -rf logs/*.log
	@rm -f $(COVERAGE_FILE) coverage.html
	@go clean -cache
	@echo "✅ 清理完成"