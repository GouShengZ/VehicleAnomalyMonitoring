package trigger

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zhangyuchen/AutoDataHub-monitor/pkg/models"
)

// APIClient 负责从外部API获取负面触发器数据
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient 创建一个新的APIClient实例
func NewAPIClient(baseURL string, timeout time.Duration) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// FetchNegativeTriggerData 从API获取负面触发器数据，支持按时间范围和触发器ID列表过滤
func (c *APIClient) FetchNegativeTriggerData(endpoint string, startTime, endTime int64, triggerIDs []string) ([]*models.NegativeTriggerData, error) {
	// 构建带查询参数的URL
	baseURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	url := baseURL

	// 添加查询参数
	queryParams := make([]string, 0)
	if startTime > 0 {
		queryParams = append(queryParams, fmt.Sprintf("start_time=%d", startTime))
	}
	if endTime > 0 {
		queryParams = append(queryParams, fmt.Sprintf("end_time=%d", endTime))
	}
	if len(triggerIDs) > 0 {
		// 将触发器ID列表转换为逗号分隔的字符串
		triggerIDsStr := ""
		for i, id := range triggerIDs {
			if i > 0 {
				triggerIDsStr += ","
			}
			triggerIDsStr += id
		}
		queryParams = append(queryParams, fmt.Sprintf("trigger_ids=%s", triggerIDsStr))
	}

	// 将查询参数添加到URL
	if len(queryParams) > 0 {
		url = fmt.Sprintf("%s?%s", baseURL, strings.Join(queryParams, "&"))
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("API请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回非成功状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 解析API响应，现在期望返回一个数组
	var apiResponses []struct {
		VIN       string `json:"id"`         // API返回的车辆ID字段
		Timestamp int64  `json:"timestamp"`  // 时间戳
		TriggerID string `json:"trigger_id"` // 触发器ID（可选）
		// Speed     int    `json:"speed"`     // 速度（可选，用于判断是否为负面触发）
	}

	if err := json.Unmarshal(body, &apiResponses); err != nil {
		return nil, fmt.Errorf("解析API响应失败: %w", err)
	}

	// 创建负面触发器数据切片
	triggerDataList := make([]*models.NegativeTriggerData, 0, len(apiResponses))
	for _, resp := range apiResponses {
		triggerData := &models.NegativeTriggerData{
			VIN:       resp.VIN,
			Timestamp: resp.Timestamp,
			TriggerID: uuid.New().String(),
			Type:      "negative",
			Status:    "pending",
		}
		triggerDataList = append(triggerDataList, triggerData)
	}

	return triggerDataList, nil
}
