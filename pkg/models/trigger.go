package models

import (
	"context"
	"encoding/json"
	"fmt"

	"AutoDataHub-monitor/configs"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// 获取Redis客户端实例
var redisClient = configs.Client.Redis // Use initialized Redis client instance

// NegativeTriggerData 表示负面触发器数据
type NegativeTriggerData struct {
	Vin       string `json:"vin"`        // 车辆识别号
	Timestamp int64  `json:"timestamp"`  // 触发时间戳
	CarType   string `json:"car_type"`   // 车辆类型
	UsageType string `json:"usage_type"` // 使用类型
	TriggerID string `json:"trigger_id"` // 触发器ID
	LogId     int    `json:"log_id"`     // 日志ID
}

// PopToRedisQueue 从指定的Redis队列中弹出一个负面触发器数据。
// 它接收一个上下文和一个队列名称作为参数。
// 如果队列为空，则返回 (nil, nil)。
// 如果发生错误（例如，Redis连接问题或反序列化失败），则返回错误。
func PopToRedisQueue(ctx context.Context, queueName string) (*NegativeTriggerData, error) {
	val, err := redisClient.LPop(ctx, queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 队列为空
		}
		return nil, fmt.Errorf("从队列获取数据失败: %v", err)
	}
	// 反序列化数据
	var data NegativeTriggerData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("反序列化数据失败: %v", err)
	}
	return &data, nil
}

// PushToRedisQueue 将触发器数据推送到Redis队列
// 它首先检查 LogId 是否为 0，如果是，则创建一个新的流程日志条目。
// 否则，它会更新现有的日志条目。
// 然后，它将数据序列化为 JSON 并将其推送到指定的 Redis 队列。
// 如果在任何步骤中发生错误，它将记录错误并返回。
func (d *NegativeTriggerData) PushToRedisQueue(queueName string) error {
	db := configs.Client.MySQL
	var err error
	if d.LogId == 0 {
		insertData := ProcessLogs{
			Vin:              d.Vin,
			TriggerTimestamp: d.Timestamp,
			CarType:          d.CarType,
			UseType:          d.UsageType,
			TriggerID:        d.TriggerID,
			ProcessStatus:    queueName + "_start",
			ProcessLog:       queueName,
		}
		var res *ProcessLogs
		res, err = CreateProcessLog(db, insertData)
		if err != nil {
			configs.Client.Logger.Error("创建处理日志失败", zap.Error(err))
			return fmt.Errorf("创建处理日志失败: %w", err)
		}
		d.LogId = res.ID
	} else {
		err = UpdateProcessLog(db, map[string]interface{}{"id": d.LogId, "process_status": queueName})
		if err != nil {
			configs.Client.Logger.Error("更新处理日志状态失败", zap.Error(err))
			return fmt.Errorf("更新处理日志状态失败: %w", err)
		}
		err = AddProcessLog(db, d.LogId, queueName)
		if err != nil {
			configs.Client.Logger.Error("添加处理日志失败", zap.Error(err))
			return fmt.Errorf("添加处理日志失败: %w", err)
		}
	}

	ctx := context.Background()

	// 将数据序列化为JSON
	jsonData, err := json.Marshal(d)
	if err != nil {
		configs.Client.Logger.Error("序列化数据失败", zap.Error(err)) // Use initialized Logger instance
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	// 推送到Redis列表
	result := redisClient.RPush(ctx, queueName, jsonData)
	if result.Err() != nil {
		configs.Client.Logger.Error("推送数据到Redis队列失败", zap.Error(result.Err())) // Use initialized Logger instance
		return fmt.Errorf("推送数据到Redis队列失败: %w", result.Err())
	}

	configs.Client.Logger.Info("成功推送数据到队列", // Use initialized Logger instance
		zap.String("queue", queueName),
		zap.String("vin", d.Vin),
		zap.Int64("timestamp", d.Timestamp))
	return nil
}

// PushToDefaultQueue 将触发器数据推送到默认队列
// 它从配置中获取默认队列名称，然后调用 PushToRedisQueue。
func (d *NegativeTriggerData) PushToDefaultQueue() error {
	// 从配置中获取默认队列名称
	queueName := configs.Cfg.VehicleType.DefaultQueue
	return d.PushToRedisQueue(queueName)
}
