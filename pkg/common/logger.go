package common

import (
	"github.com/zhangyuchen/AutoDataHub-monitor/configs"
	"go.uber.org/zap"
)

// Logger 是全局日志实例
var Logger *zap.Logger

// InitLogger 初始化全局日志实例
func InitLogger() {
	Logger = configs.GetLogger()
}

// 在包初始化时自动初始化Logger
func init() {
	InitLogger()
}
