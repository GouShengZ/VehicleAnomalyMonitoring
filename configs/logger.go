package configs

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerConfig 表示日志的配置信息
type LoggerConfig struct {
	Level      string `yaml:"level"`       // 日志级别：debug, info, warn, error, fatal
	FilePath   string `yaml:"file_path"`   // 日志文件路径
	MaxSize    int    `yaml:"max_size"`    // 单个日志文件最大大小（MB）
	MaxBackups int    `yaml:"max_backups"` // 最大保留的旧日志文件数量
	MaxAge     int    `yaml:"max_age"`     // 日志文件保留的最大天数
	Compress   bool   `yaml:"compress"`    // 是否压缩旧日志文件
	Console    bool   `yaml:"console"`     // 是否同时输出到控制台
}

// 全局日志实例
var (
	logger     *zap.Logger
	loggerOnce sync.Once
)

// DefaultLoggerConfig 返回默认的日志配置
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:      "info",
		FilePath:   "logs/app.log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
		Console:    true,
	}
}

// GetLogger 获取全局日志实例
func GetLogger() *zap.Logger {
	loggerOnce.Do(func() {
		// 确保配置已加载
		logConfig := GetConfig().Logger
		if logConfig == nil {
			logConfig = DefaultLoggerConfig()
		}

		// 解析日志级别
		var level zapcore.Level
		switch logConfig.Level {
		case "debug":
			level = zapcore.DebugLevel
		case "info":
			level = zapcore.InfoLevel
		case "warn":
			level = zapcore.WarnLevel
		case "error":
			level = zapcore.ErrorLevel
		case "fatal":
			level = zapcore.FatalLevel
		default:
			level = zapcore.InfoLevel
		}

		// 创建编码器配置
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		// 创建核心
		var cores []zapcore.Core

		// 文件输出
		if logConfig.FilePath != "" {
			// 确保日志目录存在
			lastSlashIndex := -1
			for i := len(logConfig.FilePath) - 1; i >= 0; i-- {
				if logConfig.FilePath[i] == '/' {
					lastSlashIndex = i
					break
				}
			}

			if lastSlashIndex > 0 {
				logDir := logConfig.FilePath[:lastSlashIndex]
				os.MkdirAll(logDir, 0755)
			}

			// 创建文件同步器
			fileSyncer, _ := os.Create(logConfig.FilePath)
			fileCore := zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderConfig),
				zapcore.AddSync(fileSyncer),
				level,
			)
			cores = append(cores, fileCore)
		}

		// 控制台输出
		if logConfig.Console {
			consoleCore := zapcore.NewCore(
				zapcore.NewConsoleEncoder(encoderConfig),
				zapcore.AddSync(os.Stdout),
				level,
			)
			cores = append(cores, consoleCore)
		}

		// 创建日志实例
		core := zapcore.NewTee(cores...)
		logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	})

	return logger
}

// CloseLogger 关闭日志实例
func CloseLogger() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}
