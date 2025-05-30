package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"AutoDataHub-monitor/configs"

	"go.uber.org/zap"
)

// HealthStatus 健康状态枚举
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// HealthCheck 健康检查结果
type HealthCheck struct {
	Service   string       `json:"service"`
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	Duration  string       `json:"duration"`
}

// OverallHealth 整体健康状况
type OverallHealth struct {
	Status    HealthStatus  `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Checks    []HealthCheck `json:"checks"`
	Uptime    string        `json:"uptime"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	logger    *zap.Logger
	startTime time.Time
}

// NewHealthChecker 创建新的健康检查器
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		logger:    configs.Client.Logger,
		startTime: time.Now(),
	}
}

// CheckRedis 检查Redis连接健康状况
func (h *HealthChecker) CheckRedis(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Service:   "redis",
		Timestamp: start,
	}

	if configs.Client.Redis == nil {
		check.Status = StatusUnhealthy
		check.Message = "Redis client not initialized"
		check.Duration = time.Since(start).String()
		return check
	}

	// 测试连接
	_, err := configs.Client.Redis.Ping(ctx).Result()
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Redis ping failed: %v", err)
	} else {
		check.Status = StatusHealthy
		check.Message = "Redis connection is healthy"
	}

	check.Duration = time.Since(start).String()
	return check
}

// CheckMySQL 检查MySQL连接健康状况
func (h *HealthChecker) CheckMySQL(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Service:   "mysql",
		Timestamp: start,
	}

	if configs.Client.MySQL == nil {
		check.Status = StatusUnhealthy
		check.Message = "MySQL client not initialized"
		check.Duration = time.Since(start).String()
		return check
	}

	// 测试连接
	sqlDB, err := configs.Client.MySQL.DB()
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Failed to get database instance: %v", err)
		check.Duration = time.Since(start).String()
		return check
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("MySQL ping failed: %v", err)
	} else {
		check.Status = StatusHealthy
		check.Message = "MySQL connection is healthy"
	}

	check.Duration = time.Since(start).String()
	return check
}

// StartHealthServer 启动健康检查HTTP服务器
func (h *HealthChecker) StartHealthServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.HealthHandler)
	mux.HandleFunc("/health/ready", h.HealthHandler) // Kubernetes readiness probe
	mux.HandleFunc("/health/live", h.HealthHandler)  // Kubernetes liveness probe

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	h.logger.Info("Starting health check server", zap.String("port", port))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		h.logger.Fatal("Health server failed to start", zap.Error(err))
	}
}

// HealthHandler HTTP健康检查处理器
func (h *HealthChecker) HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	health := h.GetOverallHealth(ctx)

	w.Header().Set("Content-Type", "application/json")

	// 根据健康状态设置HTTP状态码
	switch health.Status {
	case StatusHealthy:
		w.WriteHeader(http.StatusOK)
	case StatusDegraded:
		w.WriteHeader(http.StatusOK) // 降级但仍可服务
	case StatusUnhealthy:
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(health); err != nil {
		h.logger.Error("Failed to encode health response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetOverallHealth 获取整体健康状况
func (h *HealthChecker) GetOverallHealth(ctx context.Context) OverallHealth {
	checks := []HealthCheck{
		h.CheckRedis(ctx),
		h.CheckMySQL(ctx),
	}

	// 确定整体状态
	overallStatus := StatusHealthy
	for _, check := range checks {
		if check.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
			break
		} else if check.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	return OverallHealth{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Checks:    checks,
		Uptime:    time.Since(h.startTime).String(),
	}
}
