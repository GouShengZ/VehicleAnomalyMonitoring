package trigger

import (
	"fmt"
	"time"

	"github.com/zhangyuchen/AutoDataHub-monitor/configs"
	"github.com/zhangyuchen/AutoDataHub-monitor/pkg/common"

	"go.uber.org/zap"
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
		common.Logger.Error("获取负面触发器数据失败", zap.Error(err))
		return fmt.Errorf("获取负面触发器数据失败: %w", err)
	}

	// 检查是否有数据返回
	if len(triggerDataList) == 0 {
		common.Logger.Info("没有找到符合条件的负面触发器数据")
		return nil
	}

	// 将数据推送到Redis队列
	for _, data := range triggerDataList {
		err := data.PushToRedisQueue(configs.InitQueueName)
		if err != nil {
			common.Logger.Error("推送负面触发器数据到Redis失败", zap.Error(err))
		}
	}

	common.Logger.Info("成功处理负面触发器数据", zap.Int("count", len(triggerDataList)))
	return nil
}
