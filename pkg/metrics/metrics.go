package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Metrics 监控指标结构
type Metrics struct {
	// 消息处理指标
	MessagesProcessed *prometheus.CounterVec
	MessageDuration   *prometheus.HistogramVec
	MessageErrors     *prometheus.CounterVec

	// 队列指标
	QueueSize     *prometheus.GaugeVec
	QueueWaitTime *prometheus.HistogramVec

	// 系统指标
	ActiveWorkers *prometheus.GaugeVec
	SystemUptime  prometheus.Gauge

	// 数据库指标
	DBConnections   *prometheus.GaugeVec
	DBQueryDuration *prometheus.HistogramVec
	DBErrors        *prometheus.CounterVec

	// Redis 指标
	RedisConnections *prometheus.GaugeVec
	RedisLatency     *prometheus.HistogramVec
	RedisErrors      *prometheus.CounterVec
}

// NewMetrics 创建新的监控指标实例
func NewMetrics() *Metrics {
	return &Metrics{
		MessagesProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "autodatahub_messages_processed_total",
				Help: "处理的消息总数",
			},
			[]string{"queue_type", "vehicle_type", "status"},
		),
		MessageDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "autodatahub_message_duration_seconds",
				Help:    "消息处理持续时间",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"queue_type", "vehicle_type"},
		),
		MessageErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "autodatahub_message_errors_total",
				Help: "消息处理错误总数",
			},
			[]string{"queue_type", "vehicle_type", "error_type"},
		),
		QueueSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "autodatahub_queue_size",
				Help: "队列当前大小",
			},
			[]string{"queue_type", "vehicle_type"},
		),
		QueueWaitTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "autodatahub_queue_wait_time_seconds",
				Help:    "消息在队列中的等待时间",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"queue_type", "vehicle_type"},
		),
		ActiveWorkers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "autodatahub_active_workers",
				Help: "活跃工作者数量",
			},
			[]string{"worker_type"},
		),
		SystemUptime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "autodatahub_uptime_seconds",
				Help: "系统运行时间（秒）",
			},
		),
		DBConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "autodatahub_db_connections",
				Help: "数据库连接数",
			},
			[]string{"database", "status"},
		),
		DBQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "autodatahub_db_query_duration_seconds",
				Help:    "数据库查询持续时间",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"database", "operation"},
		),
		DBErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "autodatahub_db_errors_total",
				Help: "数据库错误总数",
			},
			[]string{"database", "error_type"},
		),
		RedisConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "autodatahub_redis_connections",
				Help: "Redis连接数",
			},
			[]string{"status"},
		),
		RedisLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "autodatahub_redis_latency_seconds",
				Help:    "Redis操作延迟",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		RedisErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "autodatahub_redis_errors_total",
				Help: "Redis错误总数",
			},
			[]string{"error_type"},
		),
	}
}

// Register 注册所有指标到Prometheus
func (m *Metrics) Register() error {
	metrics := []prometheus.Collector{
		m.MessagesProcessed,
		m.MessageDuration,
		m.MessageErrors,
		m.QueueSize,
		m.QueueWaitTime,
		m.ActiveWorkers,
		m.SystemUptime,
		m.DBConnections,
		m.DBQueryDuration,
		m.DBErrors,
		m.RedisConnections,
		m.RedisLatency,
		m.RedisErrors,
	}

	for _, metric := range metrics {
		if err := prometheus.Register(metric); err != nil {
			return err
		}
	}

	return nil
}

// RecordMessageProcessed 记录消息处理指标
func (m *Metrics) RecordMessageProcessed(queueType, vehicleType, status string, duration time.Duration) {
	m.MessagesProcessed.WithLabelValues(queueType, vehicleType, status).Inc()
	m.MessageDuration.WithLabelValues(queueType, vehicleType).Observe(duration.Seconds())
}

// RecordMessageError 记录消息处理错误
func (m *Metrics) RecordMessageError(queueType, vehicleType, errorType string) {
	m.MessageErrors.WithLabelValues(queueType, vehicleType, errorType).Inc()
}

// UpdateQueueSize 更新队列大小
func (m *Metrics) UpdateQueueSize(queueType, vehicleType string, size int) {
	m.QueueSize.WithLabelValues(queueType, vehicleType).Set(float64(size))
}

// RecordQueueWaitTime 记录队列等待时间
func (m *Metrics) RecordQueueWaitTime(queueType, vehicleType string, waitTime time.Duration) {
	m.QueueWaitTime.WithLabelValues(queueType, vehicleType).Observe(waitTime.Seconds())
}

// UpdateActiveWorkers 更新活跃工作者数量
func (m *Metrics) UpdateActiveWorkers(workerType string, count int) {
	m.ActiveWorkers.WithLabelValues(workerType).Set(float64(count))
}

// StartUptimeCounter 启动运行时间计数器
func (m *Metrics) StartUptimeCounter() {
	startTime := time.Now()
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for range ticker.C {
			uptime := time.Since(startTime).Seconds()
			m.SystemUptime.Set(uptime)
		}
	}()
}

// RecordDBOperation 记录数据库操作指标
func (m *Metrics) RecordDBOperation(database, operation string, duration time.Duration, err error) {
	m.DBQueryDuration.WithLabelValues(database, operation).Observe(duration.Seconds())
	if err != nil {
		m.DBErrors.WithLabelValues(database, "query_error").Inc()
	}
}

// UpdateDBConnections 更新数据库连接数
func (m *Metrics) UpdateDBConnections(database, status string, count int) {
	m.DBConnections.WithLabelValues(database, status).Set(float64(count))
}

// RecordRedisOperation 记录Redis操作指标
func (m *Metrics) RecordRedisOperation(operation string, duration time.Duration, err error) {
	m.RedisLatency.WithLabelValues(operation).Observe(duration.Seconds())
	if err != nil {
		m.RedisErrors.WithLabelValues("operation_error").Inc()
	}
}

// UpdateRedisConnections 更新Redis连接数
func (m *Metrics) UpdateRedisConnections(status string, count int) {
	m.RedisConnections.WithLabelValues(status).Set(float64(count))
}

// MetricsServer 监控指标服务器
type MetricsServer struct {
	server *http.Server
	logger *zap.Logger
}

// NewMetricsServer 创建新的监控指标服务器
func NewMetricsServer(port string, logger *zap.Logger) *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return &MetricsServer{
		server: &http.Server{
			Addr:    ":" + port,
			Handler: mux,
		},
		logger: logger,
	}
}

// Start 启动监控指标服务器
func (s *MetricsServer) Start() error {
	s.logger.Info("starting metrics server", zap.String("addr", s.server.Addr))
	return s.server.ListenAndServe()
}

// Stop 停止监控指标服务器
func (s *MetricsServer) Stop(ctx context.Context) error {
	s.logger.Info("stopping metrics server")
	return s.server.Shutdown(ctx)
}

// 全局监控指标实例
var GlobalMetrics *Metrics

// InitGlobalMetrics 初始化全局监控指标
func InitGlobalMetrics() error {
	GlobalMetrics = NewMetrics()
	err := GlobalMetrics.Register()
	if err != nil {
		return err
	}

	// 启动运行时间计数器
	GlobalMetrics.StartUptimeCounter()

	return nil
}
