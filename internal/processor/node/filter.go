package node

import (
	"context"

	"AutoDataHub-monitor/configs"
	"AutoDataHub-monitor/pkg/models"
)

var logger = configs.Client.Logger

func ProcessDefaultData() {
	defer func() {
		if err := recover(); err != nil {
			logger.Sugar().Errorf("捕获到 panic：%v\n", err)
		}
	}()

	data, err := models.PopToRedisQueue(context.Background(), configs.Cfg.VehicleType.DefaultQueue)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	queueName := ""
	// 根据使用类型进行不同处理
	switch data.UsageType {
	case "production":
		// 生产车辆处理
		queueName = configs.Cfg.VehicleType.ProductionCarQueue
	case "test_drive":
		// 试驾车辆处理
		queueName = configs.Cfg.VehicleType.TestDriveCarQueue
	case "media":
		// 媒体车辆处理
		queueName = configs.Cfg.VehicleType.MediaCarQueue
	case "internal":
		// 内部车辆处理
		queueName = configs.Cfg.VehicleType.InternalCarQueue
	default:
		queueName = configs.Cfg.VehicleType.ProductionCarQueue
	}
	data.PushToRedisQueue(queueName)

	return
}
