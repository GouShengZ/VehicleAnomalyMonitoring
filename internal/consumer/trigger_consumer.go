package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/zhangyuchen/AutoDataHub-monitor/internal/models"
)

// TriggerConsumer 触发数据消费者
type TriggerConsumer struct {
	consumer sarama.ConsumerGroup
	topic    string
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewTriggerConsumer 创建新的触发数据消费者
func NewTriggerConsumer(config *models.Config) (*TriggerConsumer, error) {
	// 创建Kafka配置
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Return.Errors = true
	kafkaConfig.ClientID = config.Kafka.ClientID + "-trigger"
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	// 使用相同的Kafka集群，但使用不同的主题
	triggerTopic := config.Kafka.CanTopic + "-trigger" // 假设触发数据主题是CAN主题加上"-trigger"后缀

	// 创建消费者组
	consumerGroup, err := sarama.NewConsumerGroup(config.Kafka.Brokers, config.Kafka.GroupID+"-trigger", kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("创建触发数据消费者组失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &TriggerConsumer{
		consumer: consumerGroup,
		topic:    triggerTopic,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Start 启动触发数据消费者
func (t *TriggerConsumer) Start(handler func(data *models.TriggerData) error) {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()

		// 创建消费者处理器
		consumerHandler := &TriggerHandler{
			ready:   make(chan bool),
			handler: handler,
		}

		for {
			select {
			case <-t.ctx.Done():
				log.Println("触发数据消费者正在关闭...")
				return
			default:
				// 消费消息
				if err := t.consumer.Consume(t.ctx, []string{t.topic}, consumerHandler); err != nil {
					log.Printf("触发数据消费错误: %v\n", err)
					time.Sleep(time.Second) // 错误后等待一秒再重试
				}
				// 检查是否需要重新连接
				if t.ctx.Err() != nil {
					return
				}
				consumerHandler.ready = make(chan bool)
			}
		}
	}()

	log.Printf("触发数据消费者已启动，正在消费主题: %s\n", t.topic)
}

// Stop 停止触发数据消费者
func (t *TriggerConsumer) Stop() {
	t.cancel()
	t.wg.Wait()
	if err := t.consumer.Close(); err != nil {
		log.Printf("关闭触发数据消费者错误: %v\n", err)
	}
	log.Println("触发数据消费者已停止")
}

// TriggerHandler 触发数据处理器
type TriggerHandler struct {
	ready   chan bool
	handler func(data *models.TriggerData) error
}

// Setup 设置消费者会话
func (h *TriggerHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

// Cleanup 清理消费者会话
func (h *TriggerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 处理消费者声明
func (h *TriggerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// 解析消息
		var triggerData models.TriggerData
		if err := json.Unmarshal(message.Value, &triggerData); err != nil {
			log.Printf("解析触发数据失败: %v\n", err)
			session.MarkMessage(message, "")
			continue
		}

		// 处理消息
		if err := h.handler(&triggerData); err != nil {
			log.Printf("处理触发数据失败: %v\n", err)
		}

		// 标记消息为已处理
		session.MarkMessage(message, "")
	}

	return nil
}