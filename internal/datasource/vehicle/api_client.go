package vehicle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// VehicleTypeAPIClient 负责从外部API获取车辆类型和使用类型信息
type VehicleTypeAPIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewVehicleTypeAPIClient 创建一个新的VehicleTypeAPIClient实例
func NewVehicleTypeAPIClient(baseURL string, timeout time.Duration) *VehicleTypeAPIClient {
	return &VehicleTypeAPIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// VehicleTypeInfo 表示从API获取的车辆类型信息
type VehicleTypeInfo struct {
	VehicleType string `json:"vehicle_type"` // 车辆类型：A、B、C
	UsageType   string `json:"usage_type"`   // 使用类型：量产车、试驾车、媒体车、内部车
}

// GetVehicleTypeInfo 从API获取车辆类型和使用类型信息
func (c *VehicleTypeAPIClient) GetVehicleTypeInfo(endpoint string, vin string) (*VehicleTypeInfo, error) {
	// 构建API请求URL
	url := fmt.Sprintf("%s%s?vin=%s", c.baseURL, endpoint, vin)

	// 发送GET请求
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("API请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回非成功状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 解析API响应
	var vehicleInfo VehicleTypeInfo
	if err := json.Unmarshal(body, &vehicleInfo); err != nil {
		return nil, fmt.Errorf("解析API响应失败: %w", err)
	}

	return &vehicleInfo, nil
}

// GetVehicleTypeInfoBatch 批量获取车辆类型和使用类型信息
func (c *VehicleTypeAPIClient) GetVehicleTypeInfoBatch(endpoint string, vins []string) (map[string]*VehicleTypeInfo, error) {
	// 构建API请求URL，支持批量查询
	url := fmt.Sprintf("%s%s/batch", c.baseURL, endpoint)

	// 构建请求体
	reqBody, err := json.Marshal(map[string][]string{"vins": vins})
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建POST请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回非成功状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 解析API响应
	var vehicleInfoMap map[string]*VehicleTypeInfo
	if err := json.Unmarshal(body, &vehicleInfoMap); err != nil {
		return nil, fmt.Errorf("解析API响应失败: %w", err)
	}

	return vehicleInfoMap, nil
}
