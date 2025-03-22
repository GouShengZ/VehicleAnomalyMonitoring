package trigger

import (
	"fmt"
	"log"
	"time"

	"github.com/zhangyuchen/AutoDataHub-monitor/configs"
	// "github.com/zhangyuchen/AutoDataHub-monitor/pkg/common"
)

// 使用configs包中的TriggerConfig

// Service 负责协调API客户端和数据处理
type Service struct {
	config    *configs.TriggerConfig
	apiClient *APIClient
}

// NewService 创建一个新的触发器服务
func NewService(config *configs.TriggerConfig, queueName string) (*Service, error) {
	// 创建API客户端
	apiClient := NewAPIClient(
		config.APIBaseURL,
		time.Duration(config.APITimeout)*time.Second,
	)

	return &Service{
		config:    config,
		apiClient: apiClient,
	}, nil
}

// FetchAndProcessNegativeTrigger 获取并处理负面触发器数据
func (s *Service) FetchAndProcessNegativeTrigger(startTime, endTime int64, triggerIDs []string) error {
	// 从API获取数据
	triggerDataList, err := s.apiClient.FetchNegativeTriggerData(s.config.APIEndpoint, startTime, endTime, triggerIDs)
	if err != nil {
		return fmt.Errorf("获取负面触发器数据失败: %w", err)
	}

	// 检查是否有数据返回
	if len(triggerDataList) == 0 {
		log.Println("没有找到符合条件的负面触发器数据")
		return nil
	}

	// 将数据推送到Redis队列
	for _, data := range triggerDataList {
		err := data.PushToRedisQueue("")
		if err != nil {
			fmt.Errorf("推送负面触发器数据到Redis失败: %w", err)
		}
	}

	log.Printf("成功处理%d条负面触发器数据", len(triggerDataList))
	return nil
}
