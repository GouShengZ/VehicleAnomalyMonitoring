package configs

// RedisConfig 表示Redis的配置信息
type RedisConfig struct {
	Addr     string `yaml:"addr"`     // Redis服务器地址
	Password string `yaml:"password"` // Redis密码
	DB       int    `yaml:"db"`       // Redis数据库编号
}

// DefaultRedisConfig 返回默认的Redis配置
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
}
