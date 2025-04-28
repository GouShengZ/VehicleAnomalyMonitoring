package configs

import (
	"fmt"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Redis       RedisConfig       `yaml:"redis"`
	Trigger     TriggerConfig     `yaml:"trigger"`
	Logger      LoggerConfig      `yaml:"logger"`
	MySQL       MySQLConfig       `yaml:"mysql"`
	VehicleType VehicleTypeConfig `yaml:"vehicle_type"`
}

type RedisConfig struct {
	Addr      string `yaml:"addr"`
	Password  string `yaml:"password"`
	DB        int    `yaml:"db"`
	QueueName string `yaml:"queue_name"`
}

type TriggerConfig struct {
	APIBaseURL         string `yaml:"api_base_url"`
	FromPath           string `yaml:"from_path"`
	FromPathMethod     string `yaml:"from_path_method"`
	TriggerIdList      []int  `yaml:"trigger_id_list"`
	DownloadPath       string `yaml:"download_path"`
	DownloadPathMethod string `yaml:"download_path_method"`
}

type LoggerConfig struct {
	Level      string `yaml:"level"`
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
	Console    bool   `yaml:"console"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
}

type VehicleTypeConfig struct {
	DefaultQueue       string `yaml:"default_queue"`
	ProductionCarQueue string `yaml:"production_car_queue"`
	TestDriveCarQueue  string `yaml:"test_drive_car_queue"`
	MediaCarQueue      string `yaml:"media_car_queue"`
	InternalCarQueue   string `yaml:"internal_car_queue"`
}

func (c *VehicleTypeConfig) ForEach(handler func(fieldName, value string)) {
	// 使用反射遍历并执行处理函数
	val := reflect.ValueOf(*c)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Tag.Get("yaml") // 获取yaml标签名
		fieldValue := field.String()
		handler(fieldName, fieldValue)
	}
}

// LoadConfig 加载YAML配置文件并返回配置结构体指针
// LoadConfig 安全加载配置文件
// 参数:
// path string - 配置文件路径
// 返回:
// *Config - 配置对象指针
// error - 加载错误信息
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var newCfg Config
	if err := yaml.Unmarshal(data, &newCfg); err != nil {
		return nil, fmt.Errorf("解析YAML配置失败: %w", err)
	}
	return &newCfg, nil
}

// LoadConfigFromFile 显式加载配置文件
// 参数:
// path string - 配置文件路径
// 返回:
// *Config - 配置对象指针
// error - 加载错误信息
func LoadConfigFromFile(path string) (*Config, error) {
	cfg, err := LoadConfig(path)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
