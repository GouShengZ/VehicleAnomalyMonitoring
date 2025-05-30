package utils

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestRetrySuccess(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 3,
		Delay:       10 * time.Millisecond,
		Backoff:     false,
		Logger:      zap.NewNop(),
	}

	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		if callCount < 2 {
			return errors.New("temporary failure")
		}
		return nil
	}

	err := Do(context.Background(), config, operation)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}

func TestRetryFailure(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 3,
		Delay:       10 * time.Millisecond,
		Backoff:     false,
		Logger:      zap.NewNop(),
	}

	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		return errors.New("persistent failure")
	}

	err := Do(context.Background(), config, operation)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestRetryBackoff(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 3,
		Delay:       10 * time.Millisecond,
		Backoff:     true,
		Logger:      zap.NewNop(),
	}

	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		return errors.New("failure")
	}

	start := time.Now()
	err := Do(context.Background(), config, operation)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}

	// 验证指数退避使得总时间比线性延迟更长
	expectedMinTime := 10 * time.Millisecond
	if duration < expectedMinTime {
		t.Errorf("Expected at least %v with backoff, got %v", expectedMinTime, duration)
	}
}

func TestRetryContextCancellation(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 5,
		Delay:       100 * time.Millisecond,
		Backoff:     false,
		Logger:      zap.NewNop(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		return errors.New("failure")
	}

	err := Do(ctx, config, operation)

	if err == nil {
		t.Error("Expected error due to context cancellation")
	}

	// 应该在超时前就返回，不会达到最大重试次数
	if callCount >= 5 {
		t.Errorf("Expected fewer than 5 calls due to timeout, got %d", callCount)
	}
}

func BenchmarkRetry(b *testing.B) {
	config := RetryConfig{
		MaxAttempts: 1,
		Delay:       time.Microsecond,
		Backoff:     false,
		Logger:      zap.NewNop(),
	}

	operation := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Do(context.Background(), config, operation)
	}
}
