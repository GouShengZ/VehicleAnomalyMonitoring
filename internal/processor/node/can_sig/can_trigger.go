package node

import (
	"encoding/json"
	"fmt"
	"os"

	"AutoDataHub-monitor/configs"
	"AutoDataHub-monitor/pkg/utils"

	"gopkg.in/yaml.v3"
)

// var logger = configs.Client.Logger

type TriggerFileData struct {
	Code    string `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Vin       string   `json:"vin"`
		Timestamp string   `json:"timestamp"`
		CarType   string   `json:"car_type"`
		Tag       []string `json:"tag"`
		FileInfo  []struct {
			FileTpye string `json:"file_type"`
			FileUrl  string `json:"file_url"`
		} `json:"file_info"`
	} `json:"data"`
}

// SignalThreshold 定义了单个信号及其阈值
type SignalThreshold struct {
	Name       string  `yaml:"name"`        // 信号名称
	SignalName string  `yaml:"signal_name"` // 信号 ID
	Threshold  float64 `yaml:"threshold"`   // 信号阈值
}

// CanSignalConfig 定义了 can_sig.yaml 文件的结构
type CanSignalConfig struct {
	Signals []SignalThreshold `yaml:"signals"` // 信号列表
}

// AngleDataPoint 存储方向盘转角数据点及其时间戳
type AngleDataPoint struct {
	Timestamp int64
	Value     float64
}

type TriggeFileFromClient struct {
	url        string
	method     string
	config     *CanSignalConfig // CAN 信号配置
	signalList []string         // 存储信号名称列表
}

// NewTriggerFromClient 创建一个新的 TriggeFileFromClient 实例
func NewTriggeFileFromClient(dbcPath string) *TriggeFileFromClient {
	client := &TriggeFileFromClient{}
	client.url = configs.Cfg.Trigger.APIBaseURL + configs.Cfg.Trigger.DownloadPath // Use correct config field names
	client.method = configs.Cfg.Trigger.DownloadPathMethod                         // Use correct config field names
	cfg, err := loadCanSignalConfig(dbcPath)
	if err != nil {
		return nil
	}
	client.config = cfg
	signalList := make([]string, 0)
	for _, signal := range cfg.Signals {
		signalList = append(signalList, signal.SignalName)
	}
	client.signalList = signalList
	return client
}

func loadCanSignalConfig(path string) (*CanSignalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 CAN 信号配置文件失败 '%s': %w", path, err)
	}

	var cfg CanSignalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析 CAN 信号 YAML 配置失败: %w", err)
	}
	return &cfg, nil
}

func (t *TriggeFileFromClient) GetCanFile(path, vin string, ts int64) (outPath string, err error) {
	var response TriggerFileData
	requestData := struct {
		Vin       string `json:"vin"`
		Timestamp int64  `json:"timestamp"`
	}{
		Vin:       vin,
		Timestamp: ts,
	}

	requestBodyBytes, err := json.Marshal(requestData)
	if err != nil {
		logger.Error(fmt.Sprintf("JSON编码失败: %v", err))
		return
	}
	if err = utils.CallAPI(t.url, t.method, nil, requestBodyBytes, &response); err != nil {
		logger.Error(fmt.Sprintf("调用API失败: %v", err))
		return
	}
	canUrl := ""
	for _, fileInfo := range response.Data.FileInfo {
		if fileInfo.FileTpye == "can" {
			canUrl = fileInfo.FileUrl
			break
		}
	}
	if err = utils.DownloadFile(canUrl, path); err != nil {
		logger.Error(fmt.Sprintf("下载文件失败: %v", err))
		return
	}
	outPath = path
	return
}
func (t *TriggeFileFromClient) GetSignalListFromFile(path string) (sigMap map[int64]map[string]float64, tsList []int64, err error) {
	defer os.Remove(path)
	sigMap, tsList, err = utils.ParseCANLogWithDBC(path, "./configs/steering_angle.dbc", t.signalList)
	return
}

func (t *TriggeFileFromClient) IsSignalsReachesThreshold(sigMap map[int64]map[string]float64, tsList []int64) (isExceeded int, logStr string, err error) {
	for _, ts := range tsList {
		signals := sigMap[ts]
		for index, signal := range t.config.Signals {
			val := signals[signal.SignalName]
			if val > signal.Threshold {
				isExceeded = index + 1
				logStr = fmt.Sprintf("信号 %s 超过阈值 %f", signal.Name, signal.Threshold)
				return
			}
		}
	}
	return
}

func (t *TriggeFileFromClient) IsCrash(vin string, ts int64) (isCrash int, crashInfo string, err error) {
	outPath, err := t.GetCanFile(fmt.Sprintf("./logs/{vin}_{ts}.can"), vin, ts)
	if err != nil {
		logger.Error(fmt.Sprintf("获取can文件失败: %v", err))
		return
	}
	sigMap, tsList, err := t.GetSignalListFromFile(outPath)
	if err != nil {
		logger.Error(fmt.Sprintf("解析can文件失败: %v", err))
		return
	}
	isCrash, crashInfo, err = t.IsSignalsReachesThreshold(sigMap, tsList)
	if err != nil {
		logger.Error(fmt.Sprintf("判断信号是否超过阈值失败: %v", err))
		return
	}
	return
}
