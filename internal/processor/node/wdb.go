package node

import (
	"context"
	"encoding/json"
	"time"

	"AutoDataHub-monitor/configs"
	"AutoDataHub-monitor/pkg/models"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WriteDBNode 结构体定义了写入数据库节点的消费者
type WriteDBNode struct {
	RedisClient *redis.Client
	DB          *gorm.DB
	Logger      *zap.Logger
	QueueName   string
}

// NewWriteDBNode 创建一个新的 WriteDBNode 实例
// redisClient: Redis 客户端连接
// db: GORM 数据库连接
// logger: Zap 日志记录器
// 返回一个新的 WriteDBNode 实例和可能的错误
func NewWriteDBNode() (*WriteDBNode, error) {
	return &WriteDBNode{
		RedisClient: configs.Client.Redis,
		DB:          configs.Client.MySQL,
		Logger:      configs.Client.Logger,
		QueueName:   configs.Cfg.VehicleType.WriteDbQueue,
	}, nil
}

// StartConsumer 启动消费者开始监听指定 Redis 列表的消息
// ctx: 上下文，用于控制消费者生命周期
// 返回可能的错误
func (n *WriteDBNode) StartConsumer(ctx context.Context) error {
	n.Logger.Info("WriteDBNode 消费者启动，监听 Redis 列表", zap.String("queue", n.QueueName))

	go func() {
		for {
			select {
			case <-ctx.Done():
				n.Logger.Info("WriteDBNode 消费者收到停止信号，正在关闭...")
				return
			default:
				// 使用 BLPop 进行阻塞式读取，超时设为 1 秒以允许检查 ctx.Done()
				result, err := n.RedisClient.BLPop(ctx, 1*time.Second, n.QueueName).Result()
				if err != nil {
					if err == redis.Nil {
						// 超时，列表为空，继续循环
						continue
					} else if err == context.Canceled || err == context.DeadlineExceeded {
						n.Logger.Info("上下文取消，停止 Redis 监听")
						return
					}
					n.Logger.Error("从 Redis 读取消息失败", zap.String("queue", n.QueueName), zap.Error(err))
					// 可以考虑增加重试逻辑或延迟
					time.Sleep(1 * time.Second) // 简单延迟
					continue
				}

				// BLPop 返回一个包含两个元素的切片：[队列名, 消息内容]
				if len(result) == 2 {
					n.handleMessage([]byte(result[1])) // result[1] 是消息内容
				} else {
					n.Logger.Warn("收到非预期的 Redis BLPop 结果", zap.Any("result", result))
				}
			}
		}
	}()

	return nil
}

// handleMessage 处理从 Redis 列表接收到的单个消息
// msgBody: 消息内容的字节切片
func (n *WriteDBNode) handleMessage(msgBody []byte) {
	var dataLog models.NegativeTriggerData
	err := json.Unmarshal(msgBody, &dataLog)
	if err != nil {
		n.Logger.Error("无法解析消息体", zap.ByteString("body", msgBody), zap.Error(err))
		// Redis BLPOP/BRPOP 在成功读取后会自动移除消息，无需 Nack
		// 考虑记录错误或将原始消息存入错误队列
		return
	}

	// 可以在这里添加一些数据验证逻辑
	dbDataLog := models.DataLogs{
		Vin:               dataLog.Vin,
		TriggerTimestamp:  dataLog.Timestamp,
		CarType:           dataLog.CarType,
		UseType:           dataLog.UsageType,
		TriggerID:         dataLog.TriggerID,
		IsCrash:           dataLog.IsCrash,
		CrashReason:       models.CrashInfoMap[dataLog.IsCrash],
		CriterionJudgment: dataLog.ThresholdLog,
	}

	// 将数据写入数据库
	result := n.DB.Create(&dbDataLog)
	if result.Error != nil {
		n.Logger.Error("无法将数据写入数据库", zap.Any("dataLog", dataLog), zap.Error(result.Error))
		// 写入数据库失败，考虑重试或将消息移至死信队列/错误日志
		// 注意：Redis 没有内置的 Nack 或重新入队机制，需要手动实现
		// 例如，可以将失败的消息推入另一个 Redis 列表
		// n.RedisClient.LPush(ctx, n.QueueName+"_error", string(msgBody))
		return
	}

	n.Logger.Info("成功处理消息并写入数据库", zap.String("vin", dataLog.Vin), zap.String("triggerId", dataLog.TriggerID))

	// Redis BLPOP/BRPOP 成功读取即表示处理（从队列移除），无需 Ack

	// 更新处理日志状态 (如果需要)
	// 假设消息体中包含 processLogId
	var msgData map[string]interface{}
	if json.Unmarshal(msgBody, &msgData) == nil {
		if logIDVal, ok := msgData["processLogId"]; ok {
			if logID, ok := logIDVal.(float64); ok { // JSON 数字默认为 float64
				updateData := map[string]interface{}{
					"id":             int(logID),
					"process_status": "Completed",
					"process_log":    gorm.Expr("CONCAT(process_log, ?)", " -> DB Write Success"),
				}
				if err := models.UpdateProcessLog(n.DB, updateData); err != nil {
					n.Logger.Error("更新 ProcessLog 状态失败", zap.Int("logId", int(logID)), zap.Error(err))
				}
			}
		}
	}
}
