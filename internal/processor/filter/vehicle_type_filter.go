package filter

import (
	"fmt"

	"github.com/zhangyuchen/AutoDataHub-monitor/configs"
	"github.com/zhangyuchen/AutoDataHub-monitor/pkg/common"
	"github.com/zhangyuchen/AutoDataHub-monitor/pkg/models"
	"go.uber.org/zap"
)

func PushCarDataToRedis(data *models.NegativeTriggerData) (err error) {
	db := configs.GetMySQLDB()
	// 根据车辆类型和使用类型构建键值
	key := fmt.Sprintf("%s_%s", data.CarType, data.UsageType)

	// 获取默认的VehicleTypeConfig配置
	vehicleTypeConfig := configs.DefaultVehicleTypeConfig()

	// 查找对应的队列名称
	queueName, exists := vehicleTypeConfig.VehicleTypeMap[key]

	// 如果找不到匹配的队列，使用默认队列
	if !exists {
		common.Logger.Warn("未找到匹配的队列，使用默认队列", zap.String("key", key))
		queueName = vehicleTypeConfig.DefaultQueue
	}

	// 将数据推送到对应的Redis队列
	models.UpdateProcessLog(db, map[string]interface{}{"id": data.LogId, "process_status": queueName + "_start"})
	return data.PushToRedisQueue(queueName)
}
