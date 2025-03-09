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

// KafkaConsumer Kafka消费者
type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	topic    string
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewKafkaConsumer 创建新的Kafka消费者
func NewKafkaConsumer(topic string, config *models.Config) (*KafkaConsumer, error) {
	// 创建Kafka配置
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Return.Errors = true
	kafkaConfig.ClientID = config.Kafka.ClientID
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	// 创建消费者组
	consumerGroup, err := sarama.NewConsumerGroup(config.Kafka.Brokers, config.Kafka.GroupID, kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("创建Kafka消费者组失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &KafkaConsumer{
		consumer: consumerGroup,
		topic:    topic,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Start 启动消费者
func (k *KafkaConsumer) Start(handler func(data *models.CanData) error) {
	k.wg.Add(1)
	go func() {
		defer k.wg.Done()

		// 创建消费者处理器
		consumerHandler := &ConsumerHandler{
			ready:   make(chan bool),
			handler: handler,
		}

		for {
			select {
			case <-k.ctx.Done():
				log.Println("Kafka消费者正在关闭...")
				return
			default:
				// 消费消息
				if err := k.consumer.Consume(k.ctx, []string{k.topic}, consumerHandler); err != nil {
					log.Printf("Kafka消费错误: %v\n", err)
					time.Sleep(time.Second) // 错误后等待一秒再重试
				}
				// 检查是否需要重新连接
				if k.ctx.Err() != nil {
					return
				}
				consumerHandler.ready = make(chan bool)
			}
		}
	}()

	log.Printf("Kafka消费者已启动，正在消费主题: %s\n", k.topic)
}

// Stop 停止消费者
func (k *KafkaConsumer) Stop() {
	k.cancel()
	k.wg.Wait()
	if err := k.consumer.Close(); err != nil {
		log.Printf("关闭Kafka消费者错误: %v\n", err)
	}
	log.Println("Kafka消费者已停止")
}

// ConsumerHandler 消费者处理器
type ConsumerHandler struct {
	ready   chan bool
	handler func(data *models.CanData) error
}

// Setup 设置消费者会话
func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

// Cleanup 清理消费者会话
func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 处理消费者声明
func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// 解析消息
		var canData models.CanData
		if err := json.Unmarshal(message.Value, &canData); err != nil {
			log.Printf("解析CAN数据失败: %v\n", err)
			session.MarkMessage(message, "")
			continue
		}

		// 处理消息
		if err := h.handler(&canData); err != nil {
			log.Printf("处理CAN数据失败: %v\n", err)
		}

		// 标记消息为已处理
		session.MarkMessage(message, "")
	}

	return nil
}
