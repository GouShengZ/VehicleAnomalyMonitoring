package pipeline

import (
	"github.com/zhangyuchen/AutoDataHub-monitor/internal/datasource/redis"
	"github.com/zhangyuchen/AutoDataHub-monitor/internal/datasource/vehicle"
	"github.com/zhangyuchen/AutoDataHub-monitor/internal/processor/filter"
	// "gorm.io/gorm/logger"
)

type CarUserFilter struct {
	TriggerConsumer *redis.TriggerConsumer
}

func NewEnergyCarQueueFilter(triggerConsumer *redis.TriggerConsumer) *CarUserFilter {
	return &CarUserFilter{
		TriggerConsumer: triggerConsumer,
	}
}

func (c *CarUserFilter) Run() {
	apiClient := vehicle.NewVehicleTypeAPIClient("", 10)
	for {
		carData, err := c.TriggerConsumer.ConsumeMessage()
		if err != nil {
			// logger.Error("消费消息失败", err)
			return
		}
		carTpyeinfo, err := apiClient.GetVehicleTypeInfo("", carData.VIN)
		if err != nil {
			// logger.Error("获取车辆类型信息失败", err)
			return
		}
		carData.CarType = carTpyeinfo.VehicleType
		carData.UsageType = carTpyeinfo.UsageType

		// 根据车型和用途类型将这条数据推入对应的Redis队列
		err = filter.PushCarDataToRedis(carData)
		if err != nil {
			// logger.Error("推送数据到Redis失败", err)
		}
	}

}
