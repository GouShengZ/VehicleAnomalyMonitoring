package trigger

import (
	"encoding/json"
	"fmt"

	"AutoDataHub-monitor/configs"
	"AutoDataHub-monitor/pkg/utils"
	"AutoDataHub-monitor/pkg/models"
)

type TriggerApiData struct {
	Code string `json:"code"`
	Status string `json:"status"`
	Message string `json:"message"`
	Total int `json:"total"`
	Page int `json:"page"`
	PageSize int `json:"page_size"`
	Data []struct {
		Vin string `json:"vin"`
		Timestamp string `json:"timestamp"`
		CarType string `json:"car_type"`
		Tag []string `json:"tag"`
		TriggerInfo []struct {
			TriggerID int `json:"trigger_id"`
			TriggerName string `json:"trigger_name"`
		} `json:"trigger_info"`
	} `json:"data"`
}



type TriggerFromClient struct {
	url string
	method string
	redisClient *redis.Client
}
func NewTriggerFromClient() (*TriggerFromClient) {
	client := &TriggerFromClient{}
	client.redisClient = configs.Client.Redis
	client.url = configs.Cfg.api_base_url + configs.Cfg.from_path
	client.method = configs.Cfg.from_path_method
	return client
}

// GetTriggerDatasToRedisQueue 从API获取触发数据并存入Redis队列
func (t *TriggerFromClient) GetTriggerDatasToRedisQueue(useType string) error {
    defer func() {
        if err := recover(); err != nil {
            configs.Client.Logger.Errorf("捕获到 panic：%v\n", err)
        }
    }()
    // 调用API 解析API
    var response TriggerApiData
    requestData := struct {
        UseType       string `json:"useType"`
        TriggerIdList []int  `json:"triggerIdList"`
    }{
        UseType:       useType,
        TriggerIdList: configs.Cfg.TriggerIdList,
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
        triggerData.Timestamp = row.Timestamp
        triggerData.CarType = row.CarType

        if useType == "1" {
            triggerData.UsageType = row.UsageType
        } else {
            vinInfo, err := FindUseTypeOfVinAndTime(configs.Client.MySQL, row.Vin, row.Timestamp)
            if err != nil {
                triggerData.UsageType = "none"
            } else {
                triggerData.UsageType = vinInfo.UseType
            }
        }

        // 查找匹配的触发器ID
        // 使用标签来实现多层循环的退出
    	TriggerLoop:
		for _, triggerInfo := range row.TriggerInfo {
			for _, configID := range configs.Cfg.TriggerIdList {
				if triggerInfo.TriggerID == configID {
					triggerData.TriggerID = triggerInfo.TriggerID
					break TriggerLoop // 使用标签直接退出外层循环
				}
			}
		}
        if err := triggerData.PushToDefaultQueue(); err != nil {
            return fmt.Errorf("推送数据到队列失败: %w", err)
        }
    }

    return nil
}
