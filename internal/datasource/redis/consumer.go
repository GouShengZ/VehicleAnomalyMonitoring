package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/zhangyuchen/AutoDataHub-monitor/configs"
	"github.com/zhangyuchen/AutoDataHub-monitor/pkg/models"
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

	log.Printf("成功从队列 %s 消费数据: VIN=%s, 时间戳=%d", c.queueName, triggerData.VIN, triggerData.Timestamp)
	return &triggerData, nil
}

// ConsumeMessageBlocking 从队列中阻塞式消费消息，如果队列为空则等待指定时间
func (c *TriggerConsumer) ConsumeMessageBlocking(timeout time.Duration) (*models.NegativeTriggerData, error) {
	// 从Redis列表左侧阻塞弹出一条数据（BLPOP操作）
	result := c.redisClient.BLPop(c.ctx, timeout, c.queueName)
	if result.Err() == redis.Nil {
		// 超时，没有数据
		return nil, nil
	} else if result.Err() != nil {
		return nil, fmt.Errorf("从Redis队列阻塞获取数据失败: %w", result.Err())
	}

	// BLPOP返回的是一个包含两个元素的切片：[队列名, 值]
	if len(result.Val()) != 2 {
		return nil, fmt.Errorf("Redis BLPOP返回了意外的结果格式")
	}

	// 获取JSON数据（第二个元素是值）
	jsonData := result.Val()[1]

	// 反序列化JSON数据
	var triggerData models.NegativeTriggerData
	if err := json.Unmarshal([]byte(jsonData), &triggerData); err != nil {
		return nil, fmt.Errorf("反序列化数据失败: %w", err)
	}

	log.Printf("成功从队列 %s 阻塞消费数据: VIN=%s, 时间戳=%d", c.queueName, triggerData.VIN, triggerData.Timestamp)
	return &triggerData, nil
}

// ConsumeMessages 批量消费指定数量的消息
func (c *TriggerConsumer) ConsumeMessages(count int) ([]*models.NegativeTriggerData, error) {
	if count <= 0 {
		return nil, fmt.Errorf("消费消息数量必须大于0")
	}

	results := make([]*models.NegativeTriggerData, 0, count)
	for i := 0; i < count; i++ {
		data, err := c.ConsumeMessage()
		if err != nil {
			return results, err
		}
		if data == nil {
			// 队列为空，返回已获取的数据
			break
		}
		results = append(results, data)
	}

	return results, nil
}

// StartConsumer 启动一个消费者，持续消费队列中的消息并通过回调函数处理
func (c *TriggerConsumer) StartConsumer(handler func(*models.NegativeTriggerData) error, pollInterval time.Duration) {
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				log.Println("消费者已停止")
				return
			default:
				// 尝试消费消息
				data, err := c.ConsumeMessageBlocking(pollInterval)
				if err != nil {
					log.Printf("消费消息时发生错误: %v", err)
					// 出错时短暂暂停，避免CPU占用过高
					time.Sleep(time.Second)
					continue
				}

				if data == nil {
					// 没有消息，继续下一次轮询
					continue
				}

				// 处理消息
				if err := handler(data); err != nil {
					log.Printf("处理消息失败: %v", err)
				}
			}
		}
	}()
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
