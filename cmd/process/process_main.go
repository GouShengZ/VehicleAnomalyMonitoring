package process

import (
	"AutoDataHub-monitor/configs"
	"AutoDataHub-monitor/internal/pipeline"
)

func main() {
	// 初始化配置和客户端
	configs.Init()

	// 启动处理流水线
	pipeline.Run()
}
