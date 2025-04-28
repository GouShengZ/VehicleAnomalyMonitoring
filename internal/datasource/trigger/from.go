package trigger

import (
	"encoding/json"
	"fmt"
	"strconv"

	"AutoDataHub-monitor/configs"
	"AutoDataHub-monitor/pkg/models"
	"AutoDataHub-monitor/pkg/utils"

	"github.com/go-redis/redis/v8"
)

var logger = configs.Client.Logger

type TriggerApiData struct {
	Code     string `json:"code"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Data     []struct {
		Vin         string   `json:"vin"`
		Timestamp   string   `json:"timestamp"`
		CarType     string   `json:"car_type"`
		Tag         []string `json:"tag"`
		TriggerInfo []struct {
			TriggerID   int    `json:"trigger_id"`
			TriggerName string `json:"trigger_name"`
		} `json:"trigger_info"`
	} `json:"data"`
}

type TriggerFromClient struct {
	url         string
	method      string
	redisClient *redis.Client // Corrected type
}

// NewTriggerFromClient 创建一个新的 TriggerFromClient 实例
func NewTriggerFromClient() *TriggerFromClient {
	client := &TriggerFromClient{}
	client.redisClient = configs.Client.Redis
	client.url = configs.Cfg.Trigger.APIBaseURL + configs.Cfg.Trigger.FromPath // Use correct config field names
	client.method = configs.Cfg.Trigger.FromPathMethod                         // Use correct config field names
	return client
}

// GetTriggerDatasToRedisQueue 从API获取触发数据并存入Redis队列
// 它调用配置的API端点，解析响应，并将每个数据记录转换为 NegativeTriggerData 结构。
// 它处理时间戳和触发器ID的类型转换，并根据 useType 参数确定 UsageType。
// 最后，它将处理后的数据推送到默认的Redis队列中。
func (t *TriggerFromClient) GetTriggerDatasToRedisQueue(useType string) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Sugar().Errorf("捕获到 panic：%v\n", err)
		}
	}()
	// 调用API 解析API
	var response TriggerApiData
	requestData := struct {
		UseType       string `json:"useType"`
		TriggerIdList []int  `json:"triggerIdList"`
	}{
		UseType:       useType,
		TriggerIdList: configs.Cfg.Trigger.TriggerIdList,
	}

	requestBodyBytes, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %w", err)
	}

	if err := utils.CallAPI(t.url, t.method, nil, requestBodyBytes, &response); err != nil {
		return fmt.Errorf("调用API失败: %w", err)
	}

	// 构建结构体 放入队列
	for _, row := range response.Data {
		var triggerData models.NegativeTriggerData
		triggerData.Vin = row.Vin
		triggerData.CarType = row.CarType

		// Convert timestamp string to int64
		timestampInt, err := strconv.ParseInt(row.Timestamp, 10, 64)
		if err != nil {
			logger.Sugar().Warnf("无法解析时间戳 '%s' 为 int64: %v, 跳过记录 VIN: %s", row.Timestamp, err, row.Vin)
			continue // Skip this record if timestamp is invalid
		}
		triggerData.Timestamp = timestampInt

		// Determine UsageType based on useType parameter
		// Both cases now call FindUseTypeOfVinAndTime as row.UsageType is not available in API response
		vinInfo, err := models.FindUseTypeOfVinAndTime(configs.Client.MySQL, row.Vin, row.Timestamp)
		if err != nil {
			logger.Sugar().Warnf("无法找到 VIN '%s' 和时间戳 '%s' 的 UseType: %v, 设置为 'none'", row.Vin, row.Timestamp, err)
			triggerData.UsageType = "none"
		} else {
			triggerData.UsageType = vinInfo.UseType
		}

		// 查找匹配的触发器ID
		// 使用标签来实现多层循环的退出
		var foundTrigger bool
	TriggerLoop:
		for _, triggerInfo := range row.TriggerInfo {
			for _, configID := range configs.Cfg.Trigger.TriggerIdList {
				if triggerInfo.TriggerID == configID {
					triggerData.TriggerID = strconv.Itoa(triggerInfo.TriggerID) // Convert int to string
					foundTrigger = true
					break TriggerLoop // 使用标签直接退出外层循环
				}
			}
		}

		if !foundTrigger {
			logger.Sugar().Warnf("VIN '%s' 在配置的 TriggerIdList 中未找到匹配的 TriggerID, 跳过记录", row.Vin)
			continue // Skip if no matching trigger ID is found
		}

		if err := triggerData.PushToDefaultQueue(); err != nil {
			// Log the error but continue processing other records
			logger.Sugar().Errorf("推送数据到队列失败 for VIN %s: %v", triggerData.Vin, err)
			// Optionally: return fmt.Errorf("推送数据到队列失败: %w", err) // Uncomment if one failure should stop all processing
		}
	}

	return nil
}
