package models

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zhangyuchen/AutoDataHub-monitor/configs"

	"go.uber.org/zap"
)

// NegativeTriggerData 表示负面触发器数据
type NegativeTriggerData struct {
	Vin       string `json:"vin"`        // 车辆识别号
	Timestamp int64  `json:"timestamp"`  // 触发时间戳
	CarType   string `json:"car_type"`   // 车辆类型
	UsageType string `json:"usage_type"` // 使用类型
	TriggerID string `json:"trigger_id"` // 触发器ID
	LogId     int    `json:"log_id"`     // 日志ID
}

// PushToRedisQueue 将触发器数据推送到Redis队列
func (d *NegativeTriggerData) PushToRedisQueue(queueName string) error {
	db := configs.Client.MySQL
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
		res, _ := CreateProcessLog(db, insertData)
		d.LogId = res.ID
	} else {
		UpdateProcessLog(db, map[string]interface{}{"id": d.LogId, "process_status": queueName})
		AddProcessLog(db, d.LogId, queueName)
	}

	ctx := context.Background()

	// 获取Redis客户端实例
	redisClient := configs.Client.Redis // Use initialized Redis client instance

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
func (d *NegativeTriggerData) PushToDefaultQueue() error {
	// 从配置中获取默认队列名称
	queueName := configs.Cfg.VehicleType.DefaultQueue
	return d.PushToRedisQueue(queueName)
}
