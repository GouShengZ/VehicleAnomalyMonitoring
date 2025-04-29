package pipeline

import (
	"AutoDataHub-monitor/configs"
	"AutoDataHub-monitor/internal/processor/node"
	"AutoDataHub-monitor/internal/processor/node/can_sig"
	"context"
	"sync"
)

var logger = configs.Client.Logger

// Run 启动数据处理管道
func Run() {
	logger.Info("pipeline run")

	var wg sync.WaitGroup

	// 处理默认数据队列
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("default data queue run")
		for i := 0; i < 10; i++ {
			go node.ProcessDefaultData()
		}
	}()

	// 处理内部车辆队列
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("internal car queue run")
		for i := 0; i < 10; i++ {
			go can_sig.ProcessCanQueueData(configs.Cfg.VehicleType.InternalCarQueue)
		}
	}()

	// 处理媒体车辆队列
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("media car queue run")
		for i := 0; i < 10; i++ {
			go can_sig.ProcessCanQueueData(configs.Cfg.VehicleType.MediaCarQueue)
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
