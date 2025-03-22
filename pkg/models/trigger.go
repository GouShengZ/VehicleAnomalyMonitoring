package models

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zhangyuchen/AutoDataHub-monitor/configs"
	"github.com/zhangyuchen/AutoDataHub-monitor/pkg/common"
	"go.uber.org/zap"
)

// NegativeTriggerData 表示负面触发器数据
type NegativeTriggerData struct {
	VIN       string `json:"vin"`        // 车辆识别号
	Timestamp int64  `json:"timestamp"`  // 触发时间戳
	CarType   string `json:"car_type"`   // 车辆类型
	UsageType string `json:"usage_type"` // 使用类型
	TriggerID string `json:"trigger_id"` // 触发器ID
	Type      string `json:"type"`       // 触发器类型，例如"negative"
	Status    string `json:"status"`     // 触发器状态，例如"pending"
}

// PushToRedisQueue 将触发器数据推送到Redis队列
func (d *NegativeTriggerData) PushToRedisQueue(queueName string) error {
	ctx := context.Background()

	// 获取Redis客户端实例
	redisClient := configs.GetRedisClient()

	// 将数据序列化为JSON
	jsonData, err := json.Marshal(d)
	if err != nil {
		common.Logger.Error("序列化数据失败", zap.Error(err))
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	// 推送到Redis列表
	result := redisClient.RPush(ctx, queueName, jsonData)
	if result.Err() != nil {
		common.Logger.Error("推送数据到Redis队列失败", zap.Error(result.Err()))
		return fmt.Errorf("推送数据到Redis队列失败: %w", result.Err())
	}

	common.Logger.Info("成功推送数据到队列",
		zap.String("queue", queueName),
		zap.String("vin", d.VIN),
		zap.Int64("timestamp", d.Timestamp))
	return nil
}

// PushToDefaultQueue 将触发器数据推送到默认队列
// 默认队列名称格式为: negative_trigger_{car_type}_{usage_type}
func (d *NegativeTriggerData) PushToDefaultQueue() error {
	// 根据车辆类型和使用类型构建队列名称
	queueName := fmt.Sprintf("negative_trigger_%s_%s", d.CarType, d.UsageType)
	return d.PushToRedisQueue(queueName)
}
