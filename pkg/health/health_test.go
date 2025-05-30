package health

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChecker 模拟健康检查器
type MockChecker struct {
	mock.Mock
}

func (m *MockChecker) Check(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockChecker) Name() string {
	args := m.Called()
	return args.String(0)
}

func TestHealthChecker(t *testing.T) {
	t.Run("健康检查成功", func(t *testing.T) {
		mockChecker := new(MockChecker)
		mockChecker.On("Check", mock.Anything).Return(nil)
		mockChecker.On("Name").Return("test-service")

		ctx := context.Background()
		err := mockChecker.Check(ctx)

		assert.NoError(t, err)
		mockChecker.AssertExpectations(t)
	})

	t.Run("健康检查失败", func(t *testing.T) {
		mockChecker := new(MockChecker)
		mockChecker.On("Check", mock.Anything).Return(assert.AnError)
		mockChecker.On("Name").Return("test-service")

		ctx := context.Background()
		err := mockChecker.Check(ctx)

		assert.Error(t, err)
		mockChecker.AssertExpectations(t)
	})
}

func TestHealthCheck(t *testing.T) {
	t.Run("创建健康检查结果", func(t *testing.T) {
		now := time.Now()
		duration := 100 * time.Millisecond

		check := HealthCheck{
			Service:   "test-service",
			Status:    StatusHealthy,
			Message:   "Service is running normally",
			Timestamp: now,
			Duration:  duration.String(),
		}

		assert.Equal(t, "test-service", check.Service)
		assert.Equal(t, StatusHealthy, check.Status)
		assert.Equal(t, "Service is running normally", check.Message)
		assert.Equal(t, now, check.Timestamp)
		assert.Equal(t, duration.String(), check.Duration)
	})

	t.Run("健康状态枚举", func(t *testing.T) {
		assert.Equal(t, HealthStatus("healthy"), StatusHealthy)
		assert.Equal(t, HealthStatus("unhealthy"), StatusUnhealthy)
		assert.Equal(t, HealthStatus("degraded"), StatusDegraded)
	})
}
