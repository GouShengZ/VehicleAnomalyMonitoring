package task

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"AutoDataHub-monitor/configs"
	"AutoDataHub-monitor/internal/datasource/trigger"
	"AutoDataHub-monitor/internal/processor/alert"
	"AutoDataHub-monitor/pkg/utils"
)

var logger = configs.Client.Logger

func main() {
	taskManager := utils.NewTaskManager()

	triggerFromClient := trigger.NewTriggerFromClient()

	taskManager.AddTask("1", "triggerApi", 5*time.Minute, func(ctx context.Context) error {
		return triggerFromClient.GetTriggerDatasToRedisQueue("1")
	})
	taskManager.AddTask("2", "redisQueueLen", 5*time.Minute, func(ctx context.Context) error {
		return alert.CheckAllRedisQueueLength(ctx)
	})

	// 启动所有任务
	successCount, err := taskManager.StartAllTasks()
	if err != nil {
		panic(err)
	}
	logger.Sugar().Infof("成功启动任务数量:", successCount)
	// 设置优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 停止所有任务
	taskManager.StopAll()
	// log.Println("所有任务已停止")
}
