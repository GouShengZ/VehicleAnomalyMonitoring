package logger

import (
	"github.com/zhangyuchen/AutoDataHub-monitor/configs"
	"go.uber.org/zap"
)

// 在文件级别定义logger变量
var logger *zap.Logger

// 在init函数中初始化logger
func init() {
	logger = configs.GetLogger()
}

// ExampleFunction 展示如何在函数中使用文件级别的logger
func ExampleFunction() {
	logger.Info("这是一个使用文件级别logger的示例")
}

// AnotherFunction 展示在另一个函数中使用相同的logger
func AnotherFunction(message string) {
	logger.Info("收到消息", zap.String("message", message))
}
