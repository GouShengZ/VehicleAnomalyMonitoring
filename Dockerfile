# 使用官方的 Golang 镜像作为构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates

# 复制 go mod 和 sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o process cmd/process/process_main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o task cmd/task/task_main.go

# 使用轻量级的 alpine 镜像作为运行阶段
FROM alpine:latest

# 安装 ca-certificates 和 wget 用于 HTTPS 请求和健康检查
RUN apk --no-cache add ca-certificates wget

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/process .
COPY --from=builder /app/task .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 创建必要的目录并设置权限
RUN mkdir -p logs && \
    chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露端口（健康检查和监控指标服务）
EXPOSE 8080

# 默认运行处理服务
CMD ["./process"]
