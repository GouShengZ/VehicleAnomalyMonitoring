package pipeline

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"AutoDataHub-monitor/configs"
	"AutoDataHub-monitor/internal/processor/node"
	"AutoDataHub-monitor/internal/processor/node/can_sig"
	"AutoDataHub-monitor/pkg/health"

	"go.uber.org/zap"
)

var logger = configs.Client.Logger

// Run 启动数据处理管道
func Run() {
	logger.Info("数据处理管道启动")

	// 创建上下文和取消函数用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// 启动健康检查服务
	healthChecker := health.NewHealthChecker()
	go func() {
		logger.Info("启动健康检查服务", zap.String("port", "8080"))
		healthChecker.StartHealthServer("8080")
	}()

	// 启动各个处理队列的工作协程
	startWorkerPools(ctx, &wg)

	// 设置信号处理用于优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待关闭信号
	<-sigChan
	logger.Info("收到关闭信号，开始优雅关闭...")

	// 取消上下文，通知所有工作协程停止
	cancel()

	// 等待所有工作协程完成，但设置超时
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("所有工作协程已安全关闭")
	case <-time.After(30 * time.Second):
		logger.Warn("等待工作协程关闭超时，强制退出")
	}
}

// startWorkerPools 启动各个工作池
func startWorkerPools(ctx context.Context, wg *sync.WaitGroup) {
	// 处理默认数据队列
	startWorkerPool(ctx, wg, "default_queue", 10, func() {
		node.ProcessDefaultData()
	})

	// 处理内部车辆队列
	startWorkerPool(ctx, wg, "internal_car_queue", 5, func() {
		can_sig.ProcessCanQueueData(configs.Cfg.VehicleType.InternalCarQueue)
	})

	// 处理媒体车辆队列
	startWorkerPool(ctx, wg, "media_car_queue", 5, func() {
		can_sig.ProcessCanQueueData(configs.Cfg.VehicleType.MediaCarQueue)
	})

	// 处理生产车辆队列
	startWorkerPool(ctx, wg, "production_car_queue", 10, func() {
		can_sig.ProcessCanQueueData(configs.Cfg.VehicleType.ProductionCarQueue)
	})

	// 处理试驾车辆队列
	startWorkerPool(ctx, wg, "test_drive_car_queue", 5, func() {
		can_sig.ProcessCanQueueData(configs.Cfg.VehicleType.TestDriveCarQueue)
	})

	// 处理写数据库队列
	wg.Add(1)
	go func() {
		defer wg.Done()
		wdbNode, err := node.NewWriteDBNode()
		if err != nil {
			logger.Sugar().Errorf("创建写数据库节点失败: %v", err)
			return
		}

		logger.Info("写数据库队列启动")
		// 启动2个写数据库工作协程
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				logger.Sugar().Infof("写数据库工作协程 %d 启动", workerID)
				if err := wdbNode.StartConsumer(ctx); err != nil {
					logger.Sugar().Errorf("写数据库工作协程 %d 出错: %v", workerID, err)
				}
				logger.Sugar().Infof("写数据库工作协程 %d 关闭", workerID)
			}(i)
		}
	}()
}

// startWorkerPool 启动工作池
func startWorkerPool(ctx context.Context, wg *sync.WaitGroup, poolName string, workerCount int, workerFunc func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Sugar().Infof("%s 工作池启动，工作协程数: %d", poolName, workerCount)

		// 启动指定数量的工作协程
		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				logger.Sugar().Infof("%s 工作协程 %d 启动", poolName, workerID)

				// 工作循环
				for {
					select {
					case <-ctx.Done():
						logger.Sugar().Infof("%s 工作协程 %d 收到停止信号", poolName, workerID)
						return
					default:
						workerFunc()
						// 添加小延迟避免过度消耗CPU
						time.Sleep(100 * time.Millisecond)
					}
				}
			}(i)
		}
	}()

	// 处理生产车辆队列
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("production car queue run")
		for i := 0; i < 10; i++ {
			go can_sig.ProcessCanQueueData(configs.Cfg.VehicleType.ProductionCarQueue)
		}
	}()

	// 处理试驾车辆队列
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("test drive car queue run")
		for i := 0; i < 10; i++ {
			go can_sig.ProcessCanQueueData(configs.Cfg.VehicleType.TestDriveCarQueue)
		}
	}()

	// 处理感知车辆队列 TODO

	// 处理写数据库队列 TODO
	wg.Add(1)
	wdbNode, err := node.NewWriteDBNode()
	if err != nil {
		logger.Sugar().Errorf("create write db node failed", err)
		return
	}
	go func() {
		defer wg.Done()
		logger.Info("write db queue run")
		for i := 0; i < 2; i++ {
			go wdbNode.StartConsumer(context.Background())
		}
	}()
	wg.Wait()
}
