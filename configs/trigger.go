package configs

// TriggerConfig 表示触发器服务的配置
type TriggerConfig struct {
	APIBaseURL    string `yaml:"api_base_url"`   // API基础URL
	APIEndpoint   string `yaml:"api_endpoint"`   // API端点
	RedisAddr     string `yaml:"redis_addr"`     // Redis服务器地址
	RedisPassword string `yaml:"redis_password"` // Redis密码
	RedisDB       int    `yaml:"redis_db"`       // Redis数据库编号
	RedisQueue    string `yaml:"redis_queue"`    // Redis队列名称
	APITimeout    int    `yaml:"api_timeout"`    // 超时时间（秒）
}

// DefaultTriggerConfig 返回默认的触发器配置
func DefaultTriggerConfig() *TriggerConfig {
	return &TriggerConfig{
		APIBaseURL:    "http://localhost:8080",
		APIEndpoint:   "/api/triggers",
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
		RedisQueue:    "negative_triggers",
		APITimeout:    30,
	}
}
