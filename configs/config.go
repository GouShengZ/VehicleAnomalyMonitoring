package configs

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-redis/redis/v8"
	"gopkg.in/yaml.v3"
)

// AppConfig 应用程序配置结构
type AppConfig struct {
	Redis   *RedisConfig   `yaml:"redis"`
	Trigger *TriggerConfig `yaml:"trigger"`
	Logger  *LoggerConfig  `yaml:"logger"`
	MySQL   *MySQLConfig   `yaml:"mysql"`
}

// 全局配置实例
var (
	config     *AppConfig
	configOnce sync.Once

	// 全局Redis客户端实例
	redisClient     *redis.Client
	redisClientOnce sync.Once
)

// LoadConfig 从指定路径加载配置文件
func LoadConfig(configPath string) (*AppConfig, error) {
	configOnce.Do(func() {
		data, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Printf("读取配置文件失败: %v\n", err)
			// 使用默认配置
			config = &AppConfig{
				Redis:   DefaultRedisConfig(),
				Trigger: DefaultTriggerConfig(),
				Logger:  DefaultLoggerConfig(),
				MySQL:   DefaultMySQLConfig(),
			}
			return
		}

		config = &AppConfig{}
		err = yaml.Unmarshal(data, config)
		if err != nil {
			fmt.Printf("解析配置文件失败: %v\n", err)
			// 使用默认配置
			config = &AppConfig{
				Redis:   DefaultRedisConfig(),
				Trigger: DefaultTriggerConfig(),
				Logger:  DefaultLoggerConfig(),
				MySQL:   DefaultMySQLConfig(),
			}
		}
	})

	return config, nil
}

// GetConfig 获取全局配置实例
func GetConfig() *AppConfig {
	if config == nil {
		// 如果配置未初始化，尝试从默认路径加载
		_, err := LoadConfig("configs/config.yaml")
		if err != nil {
			// 使用默认配置
			configOnce.Do(func() {
				config = &AppConfig{
					Redis:   DefaultRedisConfig(),
					Trigger: DefaultTriggerConfig(),
					Logger:  DefaultLoggerConfig(),
					MySQL:   DefaultMySQLConfig(),
				}
			})
		}
	}
	return config
}

// GetRedisClient 获取全局Redis客户端实例
func GetRedisClient() *redis.Client {
	redisClientOnce.Do(func() {
		// 确保配置已加载
		redisConfig := GetConfig().Redis

		// 创建Redis客户端
		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisConfig.Addr,
			Password: redisConfig.Password,
			DB:       redisConfig.DB,
		})
	})

	return redisClient
}

// CloseRedisClient 关闭Redis客户端连接
func CloseRedisClient() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}
