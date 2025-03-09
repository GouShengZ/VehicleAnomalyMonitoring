package utils

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/zhangyuchen/AutoDataHub-monitor/internal/models"
	"gopkg.in/yaml.v3"
)

// LoadConfig 从YAML文件加载配置
// 移除本地KafkaConfig定义
// 使用models包中的KafkaConfig

func LoadConfig(configPath string) (*models.Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config models.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证配置
	if config.Kafka.GroupID == "" {
		return nil, fmt.Errorf("Kafka GroupID不能为空")
	}

	if config.DBC.Enabled && config.DBC.FilePath == "" {
		return nil, fmt.Errorf("DBC解析已启用但未指定DBC文件路径")
	}

	return &config, nil
}

type KafkaConfig struct {
    Brokers   []string `yaml:"brokers"`
    CanTopic  string   `yaml:"can_topic"`
    Topics    []string `yaml:"topics"`
    GroupID   string   `yaml:"group_id"`
}