package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts int           // 最大重试次数
	Delay       time.Duration // 重试间隔
	Backoff     bool          // 是否使用指数退避
	Logger      *zap.Logger   // 日志记录器
}

// DefaultRetryConfig 默认重试配置
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	Delay:       time.Second,
	Backoff:     true,
}

// Do 执行带重试的操作
func Do(ctx context.Context, config RetryConfig, operation func(ctx context.Context) error) error {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = DefaultRetryConfig.MaxAttempts
	}
	if config.Delay <= 0 {
		config.Delay = DefaultRetryConfig.Delay
	}

	var lastErr error
	delay := config.Delay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		if config.Logger != nil {
			config.Logger.Debug("执行操作",
				zap.Int("attempt", attempt),
				zap.Int("maxAttempts", config.MaxAttempts))
		}

		err := operation(ctx)
		if err == nil {
			return nil // 成功
		}

		lastErr = err

		// 检查是否是不可重试的错误
		if IsNonRetryableError(err) {
			if config.Logger != nil {
				config.Logger.Error("遇到不可重试错误", zap.Error(err))
			}
			return err
		}

		// 如果不是最后一次尝试，等待后重试
		if attempt < config.MaxAttempts {
			if config.Logger != nil {
				config.Logger.Warn("操作失败，准备重试",
					zap.Error(err),
					zap.Int("attempt", attempt),
					zap.Duration("delay", delay))
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}

			// 指数退避
			if config.Backoff {
				delay *= 2
			}
		}
	}

	return fmt.Errorf("操作在%d次尝试后仍然失败，最后错误: %w", config.MaxAttempts, lastErr)
}

// NonRetryableError 不可重试的错误类型
type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string {
	return fmt.Sprintf("不可重试错误: %v", e.Err)
}

func (e *NonRetryableError) Unwrap() error {
	return e.Err
}

// NewNonRetryableError 创建不可重试错误
func NewNonRetryableError(err error) error {
	return &NonRetryableError{Err: err}
}

// IsNonRetryableError 检查是否是不可重试错误
func IsNonRetryableError(err error) bool {
	var nonRetryableErr *NonRetryableError
	return errors.As(err, &nonRetryableErr)
}
