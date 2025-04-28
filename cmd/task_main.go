package autodatahub-monitor


import (
	"time"
    "context"

	"AutoDataHub-monitor/pkg/utils"
	"AutoDataHub-monitor/internal/datasource/trigger"
    "AutoDataHub-monitor/internal/processor/alert"
)


func main() {
	taskManager := utils.NewTaskManager()

	triggerFromClient = trigger.NewTriggerFromClient()


	taskManager.AddTask("1", "triggerApi", 5*time.Minute, triggerFromClient.GetTriggerDatasToRedisQueue("1"))
    taskManager.AddTask("2", "redisQueueLen", 5*time.Minute, alert.CheckAllRedisQueueLength(context.Background()))

	// 启动所有任务
    successCount, err := taskManager.StartAllTasks()
    if err != nil {
        panic(err)
    }
    // 设置优雅退出
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    // 停止所有任务
    manager.StopAll()
    // log.Println("所有任务已停止")
}