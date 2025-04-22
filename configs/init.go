package configs

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Clients 聚合所有客户端实例
type Clients struct {
	Redis  *redis.Client
	MySQL  *gorm.DB
	Logger *zap.Logger
}

var Client *Clients
var Cfg *Config

// InitAll 显式初始化所有客户端
// 参数:
// cfg *Config - 配置对象指针
// 返回:
// error - 初始化错误信息
func Init() {
	cfg, err := LoadConfigFromFile("config.yaml")
	if err != nil {
		panic(err)
	}
	Client, err = InitAllClients(cfg)
	if err != nil {
		panic(err)
	}
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
	dsn := cfg.Username + ":" + cfg.Password + "@tcp(" + cfg.Host + ":" + string(cfg.Port) + ")/" + cfg.Database + "?charset=" + cfg.Charset + "&parseTime=True&loc=Local"
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
