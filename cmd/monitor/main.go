package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangyuchen/AutoDataHub-monitor/internal/consumer"
	"github.com/zhangyuchen/AutoDataHub-monitor/internal/monitor"
	"github.com/zhangyuchen/AutoDataHub-monitor/pkg/utils"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	config, err := utils.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化监控服务
	monitorService, err := monitor.NewMonitorService(config)
	if err != nil {
		log.Fatalf("初始化监控服务失败: %v", err)
	}

	// 在线模式：初始化消费者
	canConsumer, err := consumer.NewKafkaConsumer(config.Kafka.CanTopic, config)
	if err != nil {
		log.Fatalf("初始化CAN数据消费者失败: %v", err)
	}

	triggerConsumer, err := consumer.NewTriggerConsumer(config)
	if err != nil {
		log.Fatalf("初始化Trigger数据消费者失败: %v", err)
	}

	// 启动服务
	go canConsumer.Start(monitorService.ProcessCanData)
	go triggerConsumer.Start(monitorService.ProcessTriggerData)
	go monitorService.Start()

	fmt.Println("车辆状态监控系统已启动...")

	// 初始化Gin引擎
	router := gin.Default()

	// 添加健康检查路由
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	// 添加CAN文件解析API路由
	router.POST("/api/v1/can/parse", func(c *gin.Context) {
		// 获取表单参数
		vehicleID := c.DefaultPostForm("vehicleId", "unknown")
		vehicleType := c.DefaultPostForm("vehicleType", "sedan")

		// 获取上传的文件
		file, err := c.FormFile("canFile")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("获取上传文件失败: %v", err),
			})
			return
		}

		// 创建临时文件
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("打开上传文件失败: %v", err),
			})
			return
		}
		defer src.Close()

		tempFile, err := createTempFile(src, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("创建临时文件失败: %v", err),
			})
			return
		}
		defer os.Remove(tempFile) // 处理完成后删除临时文件

		// 解析CAN文件
		if err := monitorService.ParseCANFile(tempFile, vehicleID, vehicleType); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("解析CAN文件失败: %v", err),
			})
			return
		}

		// 返回成功响应
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "CAN文件解析完成",
		})
	})

	// 启动HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Base.Port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号优雅地关闭服务
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 关闭服务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP服务器关闭异常: %v", err)
	}

	canConsumer.Stop()
	triggerConsumer.Stop()
	monitorService.Stop()

	fmt.Println("车辆状态监控系统已关闭")
}

// createTempFile 创建临时文件并写入上传的文件内容
func createTempFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	// 创建临时文件
	tempFile := filepath.Join(os.TempDir(), header.Filename)
	dst, err := os.Create(tempFile)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// 写入文件内容
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return tempFile, nil
}
