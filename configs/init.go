package configs

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Client *Clients
var Cfg *Config

// Init 初始化配置
func Init() {
	var err error
	configPath := getConfigPath()

	// 从文件加载配置
	Cfg, err = loadFromFile(configPath)
	if err != nil {
		// 如果配置加载失败，使用默认配置
		fmt.Printf("配置加载失败: %v，使用默认配置\n", err)
		Cfg = getDefaultConfig()
	}

	// 使用环境变量覆盖配置
	overrideWithEnv(Cfg)
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	// 优先使用环境变量指定的配置文件路径
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		return configPath
	}

	// 默认配置文件路径
	return "./configs/config.yaml"
}

// loadFromFile 从文件加载配置
func loadFromFile(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *Config {
	return &Config{
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		MySQL: MySQLConfig{
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "",
			Database: "autodatahub",
			Charset:  "utf8mb4",
		},
		Logger: LoggerConfig{
			Level:      "info",
			FilePath:   "logs/app.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
			Console:    true,
		},
		VehicleType: VehicleTypeConfig{
			DefaultQueue:       "default_triggers",
			ProductionCarQueue: "production_car_triggers",
			TestDriveCarQueue:  "test_drive_car_triggers",
			WriteDbQueue:       "write_db_triggers",
		},
	}
}

// overrideWithEnv 使用环境变量覆盖配置
func overrideWithEnv(config *Config) {
	// Redis配置
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		config.Redis.Addr = addr
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		config.Redis.Password = password
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		if d, err := strconv.Atoi(db); err == nil {
			config.Redis.DB = d
		}
	}

	// MySQL配置
	if host := os.Getenv("MYSQL_HOST"); host != "" {
		config.MySQL.Host = host
	}
	if port := os.Getenv("MYSQL_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.MySQL.Port = p
		}
	}
	if user := os.Getenv("MYSQL_USER"); user != "" {
		config.MySQL.Username = user
	}
	if password := os.Getenv("MYSQL_PASSWORD"); password != "" {
		config.MySQL.Password = password
	}
	if database := os.Getenv("MYSQL_DATABASE"); database != "" {
		config.MySQL.Database = database
	}

	// 日志配置
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Logger.Level = level
	}
	if filepath := os.Getenv("LOG_FILE_PATH"); filepath != "" {
		config.Logger.FilePath = filepath
	}

	// 队列配置
	if queue := os.Getenv("PRODUCTION_CAR_QUEUE"); queue != "" {
		config.VehicleType.ProductionCarQueue = queue
	}
	if queue := os.Getenv("TEST_DRIVE_CAR_QUEUE"); queue != "" {
		config.VehicleType.TestDriveCarQueue = queue
	}
	if queue := os.Getenv("WRITE_DB_QUEUE"); queue != "" {
		config.VehicleType.WriteDbQueue = queue
	}
}

// Clients 聚合所有客户端实例
type Clients struct {
	Redis  *redis.Client
	MySQL  *gorm.DB
	Logger *zap.Logger
}

// InitRedis 初始化Redis连接
func InitRedis(cfg *RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	_, err := client.Ping(context.Background()).Result()
	return client, err
}

// InitMySQL 初始化MySQL连接
func InitMySQL(cfg *MySQLConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

// InitLogger 初始化日志
func InitLogger(cfg *LoggerConfig) (*zap.Logger, error) {
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.FilePath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	})

	level := zap.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zap.DebugLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "fatal":
		level = zap.FatalLevel
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		w,
		level,
	)

	var opts []zap.Option
	if cfg.Console {
		opts = append(opts, zap.AddCaller())
	}

	return zap.New(core, opts...), nil
}

// InitAllClients 初始化所有客户端并返回聚合实例
func InitAllClients(cfg *Config) (*Clients, error) {
	redisClient, err := InitRedis(&cfg.Redis)
	if err != nil {
		return nil, err
	}

	mysqlDB, err := InitMySQL(&cfg.MySQL)
	if err != nil {
		return nil, err
	}

	logger, err := InitLogger(&cfg.Logger)
	if err != nil {
		return nil, err
	}

	return &Clients{
		Redis:  redisClient,
		MySQL:  mysqlDB,
		Logger: logger,
	}, nil
}
