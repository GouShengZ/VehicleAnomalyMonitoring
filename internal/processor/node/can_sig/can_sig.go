package can_sig

import (
	"context"
	"fmt"

	"AutoDataHub-monitor/configs"
	"AutoDataHub-monitor/pkg/models"
)

var logger = configs.Client.Logger

func ProcessCanQueueData(queueName string) {
	defer func() {
		if err := recover(); err != nil {
			logger.Sugar().Errorf("捕获到 panic：%v\n", err)
		}
	}()

	data, err := models.PopToRedisQueue(context.Background(), queueName)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	// "./configs/can_sig.yaml"
	processor := NewTriggeFileFromClient(fmt.Sprintf("./configs/can_sig_%s.yaml", queueName))
	isCrash, crashInfo, err := processor.IsCrash(data.Vin, data.Timestamp)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	if isCrash != 0 {
		logger.Info(crashInfo)
		// 推入数据库队列
		data.PushToRedisQueue(configs.Cfg.VehicleType.WriteDbQueue)
	} else {
		// 推入感知队列
		data.PushToRedisQueue(configs.Cfg.VehicleType.FusionCarQueue)
	}
	return
}
