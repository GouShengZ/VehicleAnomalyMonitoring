package models

import "time"

// VehicleType 车辆类型
type VehicleType struct {
	Name        string `yaml:"name" json:"name"`
	Code        string `yaml:"code" json:"code"`
	Description string `yaml:"description" json:"description"`
}

// Config 系统配置
type Config struct {
    Base     BaseConfig    `yaml:"base"`
    Kafka    KafkaConfig   `yaml:"kafka"`
    DBC      DBCConfig     `yaml:"dbc"`
    Monitor  MonitorConfig `yaml:"monitor"`
    Database DatabaseConfig `yaml:"database"`
}

type KafkaConfig struct {
    Brokers   []string `yaml:"brokers"`
    CanTopic  string   `yaml:"can_topic"`
    Topics    []string `yaml:"topics"`
    GroupID   string   `yaml:"group_id"`
}

// ThresholdsConfig 阈值配置
type ThresholdsConfig struct {
	Can        CanThresholds        `yaml:"can"`
	Perception PerceptionThresholds `yaml:"perception"`
}

// CanThresholds CAN信号阈值
type CanThresholds struct {
	Speed          SpeedThreshold     `yaml:"speed"`
	EngineTemp     EngineTempThreshold `yaml:"engineTemp"`
	BatteryVoltage BatteryThreshold    `yaml:"batteryVoltage"`
	FuelLevel      FuelThreshold       `yaml:"fuelLevel"`
}

// SpeedThreshold 速度阈值
type SpeedThreshold struct {
	Max float64 `yaml:"max"`
	Min float64 `yaml:"min"`
}

// EngineTempThreshold 发动机温度阈值
type EngineTempThreshold struct {
	Max     float64 `yaml:"max"`
	Warning float64 `yaml:"warning"`
}

// BatteryThreshold 电池阈值
type BatteryThreshold struct {
	Min float64 `yaml:"min"`
	Max float64 `yaml:"max"`
}

// FuelThreshold 燃油阈值
type FuelThreshold struct {
	Low float64 `yaml:"low"`
}

// PerceptionThresholds 感知数据阈值
type PerceptionThresholds struct {
	ObstacleDistance ObstacleThreshold `yaml:"obstacleDistance"`
	LaneDeviation    LaneThreshold     `yaml:"laneDeviation"`
}

// ObstacleThreshold 障碍物阈值
type ObstacleThreshold struct {
	Critical float64 `yaml:"critical"`
	Warning  float64 `yaml:"warning"`
}

// LaneThreshold 车道偏离阈值
type LaneThreshold struct {
	Max float64 `yaml:"max"`
}

// ProcessingConfig 数据处理配置
type ProcessingConfig struct {
	BufferSize int `yaml:"bufferSize"`
	Interval   int `yaml:"interval"`
}

// AlertsConfig 告警配置
type AlertsConfig struct {
	EnableEmail    bool     `yaml:"enableEmail"`
	EmailReceivers []string `yaml:"emailReceivers"`
	EnableSMS      bool     `yaml:"enableSMS"`
	SMSReceivers   []string `yaml:"smsReceivers"`
}

// DBCConfig DBC文件配置
type DBCConfig struct {
	FilePath string `yaml:"filePath"`
	Enabled  bool   `yaml:"enabled"`
}

// CanData CAN数据
type CanData struct {
	Timestamp    time.Time         `json:"timestamp"`
	VehicleID    string            `json:"vehicleId"`
	VehicleType  string            `json:"vehicleType"`
	Signals      map[string]float64 `json:"signals"`
	RawData      []byte            `json:"rawData,omitempty"`
}

// TriggerData 触发数据
type TriggerData struct {
	Timestamp    time.Time         `json:"timestamp"`
	VehicleID    string            `json:"vehicleId"`
	VehicleType  string            `json:"vehicleType"`
	TriggerType  string            `json:"triggerType"`
	TriggerValue float64           `json:"triggerValue"`
	Location     *Location         `json:"location,omitempty"`
	AdditionalData map[string]interface{} `json:"additionalData,omitempty"`
}

// Location 位置信息
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude,omitempty"`
}

// Alert 告警信息
type Alert struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	VehicleID   string    `json:"vehicleId"`
	VehicleType string    `json:"vehicleType"`
	AlertType   string    `json:"alertType"`
	Severity    string    `json:"severity"` // info, warning, error, critical
	Message     string    `json:"message"`
	SourceData  interface{} `json:"sourceData,omitempty"`
}

// DBCSignal DBC信号定义
type DBCSignal struct {
	Name      string  `json:"name"`
	StartBit  int     `json:"startBit"`
	Length    int     `json:"length"`
	Factor    float64 `json:"factor"`
	Offset    float64 `json:"offset"`
	Minimum   float64 `json:"minimum"`
	Maximum   float64 `json:"maximum"`
	Unit      string  `json:"unit"`
	IsInteger bool    `json:"isInteger"`
}

// DBCMessage DBC消息定义
type DBCMessage struct {
	ID      uint32               `json:"id"`
	Name    string               `json:"name"`
	DLC     int                  `json:"dlc"`
	Signals map[string]DBCSignal `json:"signals"`
}

type BaseConfig struct {
    Port     int    `yaml:"port"`
    LogLevel string `yaml:"log_level"`
}

type MonitorConfig struct {
    Interval       string `yaml:"interval"`
    AlertThreshold int    `yaml:"alert_threshold"`
}

type DatabaseConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    Password string `yaml:"password"`
    Name     string `yaml:"name"`
}