package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/zhangyuchen/AutoDataHub-monitor/configs"
	"github.com/zhangyuchen/AutoDataHub-monitor/pkg/common"
	"github.com/zhangyuchen/AutoDataHub-monitor/pkg/models"
	"go.uber.org/zap"
)

// TriggerConsumer 负责从Redis队列中消费触发器数据
type TriggerConsumer struct {
	redisClient *redis.Client
	queueName   string
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewTriggerConsumer 创建一个新的TriggerConsumer实例
func NewTriggerConsumer(config *configs.RedisConfig, queueName string) (*TriggerConsumer, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// 测试连接
	ctx, cancel := context.WithCancel(context.Background())
	_, err := client.Ping(ctx).Result()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("连接Redis失败: %w", err)
	}

	return &TriggerConsumer{
		redisClient: client,
		queueName:   queueName,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// ConsumeMessage 从队列中消费单条消息，如果队列为空则返回nil
func (c *TriggerConsumer) ConsumeMessage() (*models.NegativeTriggerData, error) {
	db := configs.GetMySQLDB()

	// 从Redis列表左侧弹出一条数据（LPOP操作）
	result := c.redisClient.LPop(c.ctx, c.queueName)
	if result.Err() == redis.Nil {
		// 队列为空
		return nil, nil
	} else if result.Err() != nil {
		return nil, fmt.Errorf("从Redis队列获取数据失败: %w", result.Err())
	}

	// 获取JSON数据
	jsonData := result.Val()

	// 反序列化JSON数据
	var triggerData models.NegativeTriggerData
	if err := json.Unmarshal([]byte(jsonData), &triggerData); err != nil {
		return nil, fmt.Errorf("反序列化数据失败: %w", err)
	}
	models.UpdateProcessLog(db, map[string]interface{}{"id": triggerData.LogId, "process_status": c.queueName + "_running"})
	common.Logger.Info("成功从队列消费数据",
		zap.String("queue", c.queueName),
		zap.String("vin", triggerData.Vin),
		zap.Int64("timestamp", triggerData.Timestamp))
	return &triggerData, nil
}

// StopConsumer 停止消费者
func (c *TriggerConsumer) StopConsumer() {
	c.cancel()
}

// Close 关闭Redis连接
func (c *TriggerConsumer) Close() error {
	c.cancel()
	return c.redisClient.Close()
}
