package monitor

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zhangyuchen/AutoDataHub-monitor/internal/models"
	"github.com/zhangyuchen/AutoDataHub-monitor/internal/parser"
)

// MonitorService 监控服务
type MonitorService struct {
	config       *models.Config
	canDataChan  chan *models.CanData
	triggerChan  chan *models.TriggerData
	alertChan    chan *models.Alert
	stopChan     chan struct{}
	wg           sync.WaitGroup
	vehicleTypes map[string]*models.VehicleType // 车辆类型映射，便于快速查找
	parserService *parser.ParserService        // 解析服务
}

// NewMonitorService 创建新的监控服务
func NewMonitorService(config *models.Config) (*MonitorService, error) {
	// 创建车辆类型映射
	vehicleTypes := make(map[string]*models.VehicleType)
	for i, vt := range config.VehicleTypes {
		vehicleTypes[vt.Code] = &config.VehicleTypes[i]
	}

	// 创建解析服务
	parserService, err := parser.NewParserService(config)
	if err != nil {
		return nil, fmt.Errorf("初始化解析服务失败: %v", err)
	}

	return &MonitorService{
		config:       config,
		canDataChan:  make(chan *models.CanData, config.Processing.BufferSize),
		triggerChan:  make(chan *models.TriggerData, config.Processing.BufferSize),
		alertChan:    make(chan *models.Alert, config.Processing.BufferSize),
		stopChan:     make(chan struct{}),
		vehicleTypes: vehicleTypes,
		parserService: parserService,
	}, nil
}

// Start 启动监控服务
func (m *MonitorService) Start() {
	m.wg.Add(2)

	// 启动CAN数据处理协程
	go func() {
		defer m.wg.Done()
		processingTicker := time.NewTicker(time.Duration(m.config.Processing.Interval) * time.Millisecond)
		defer processingTicker.Stop()

		for {
			select {
			case <-m.stopChan:
				log.Println("CAN数据处理协程正在关闭...")
				return
			case <-processingTicker.C:
				// 处理积累的CAN数据
				m.processAccumulatedCanData()
			}
		}
	}()

	// 启动告警处理协程
	go func() {
		defer m.wg.Done()

		for {
			select {
			case <-m.stopChan:
				log.Println("告警处理协程正在关闭...")
				return
			case alert := <-m.alertChan:
				// 处理告警
				m.processAlert(alert)
			}
		}
	}()

	log.Println("监控服务已启动")
}

// Stop 停止监控服务
func (m *MonitorService) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	log.Println("监控服务已停止")
}

// ProcessCanData 处理CAN数据
func (m *MonitorService) ProcessCanData(data *models.CanData) error {
	// 将数据放入通道
	select {
	case m.canDataChan <- data:
		// 成功放入通道
	default:
		// 通道已满，记录警告
		log.Printf("警告: CAN数据通道已满，丢弃数据: %v\n", data.VehicleID)
	}

	// 立即检查关键阈值
	m.checkCanThresholds(data)

	return nil
}

// ProcessTriggerData 处理触发数据
func (m *MonitorService) ProcessTriggerData(data *models.TriggerData) error {
	// 将数据放入通道
	select {
	case m.triggerChan <- data:
		// 成功放入通道
	default:
		// 通道已满，记录警告
		log.Printf("警告: 触发数据通道已满，丢弃数据: %v\n", data.VehicleID)
	}

	// 立即处理触发数据
	m.processTriggerData(data)

	return nil
}

// processAccumulatedCanData 处理积累的CAN数据
func (m *MonitorService) processAccumulatedCanData() {
	// 处理通道中的所有数据
	processCount := 0
	for processCount < m.config.Processing.BufferSize {
		select {
		case data := <-m.canDataChan:
			// 处理CAN数据
			m.processCanData(data)
			processCount++
		default:
			// 通道为空，退出循环
			return
		}
	}
}

// processCanData 处理单个CAN数据
func (m *MonitorService) processCanData(data *models.CanData) {
	// 检查CAN数据阈值
	m.checkCanThresholds(data)

	// 这里可以添加更多的处理逻辑
	// 例如：数据存储、统计分析等
}

// processTriggerData 处理触发数据
func (m *MonitorService) processTriggerData(data *models.TriggerData) {
	// 根据触发类型进行不同的处理
	switch data.TriggerType {
	case "obstacle":
		// 处理障碍物触发
		m.checkObstacleThresholds(data)
	case "lane_deviation":
		// 处理车道偏离触发
		m.checkLaneDeviationThresholds(data)
	default:
		log.Printf("未知的触发类型: %s\n", data.TriggerType)
	}
}

// checkCanThresholds 检查CAN数据阈值
func (m *MonitorService) checkCanThresholds(data *models.CanData) {
	// 检查速度
	if speed, ok := data.Signals["speed"]; ok {
		if speed > m.config.Thresholds.Can.Speed.Max {
			m.createAlert(data.VehicleID, data.VehicleType, "speed_exceeded", "critical",
				fmt.Sprintf("车速超过最大阈值: %.1f km/h (最大: %.1f km/h)", speed, m.config.Thresholds.Can.Speed.Max),
				data)
		}
	}

	// 检查发动机温度
	if engineTemp, ok := data.Signals["engine_temp"]; ok {
		if engineTemp > m.config.Thresholds.Can.EngineTemp.Max {
			m.createAlert(data.VehicleID, data.VehicleType, "engine_temp_critical", "critical",
				fmt.Sprintf("发动机温度过高: %.1f °C (最大: %.1f °C)", engineTemp, m.config.Thresholds.Can.EngineTemp.Max),
				data)
		} else if engineTemp > m.config.Thresholds.Can.EngineTemp.Warning {
			m.createAlert(data.VehicleID, data.VehicleType, "engine_temp_warning", "warning",
				fmt.Sprintf("发动机温度警告: %.1f °C (警告: %.1f °C)", engineTemp, m.config.Thresholds.Can.EngineTemp.Warning),
				data)
		}
	}

	// 检查电池电压
	if batteryVoltage, ok := data.Signals["battery_voltage"]; ok {
		if batteryVoltage < m.config.Thresholds.Can.BatteryVoltage.Min {
			m.createAlert(data.VehicleID, data.VehicleType, "battery_voltage_low", "warning",
				fmt.Sprintf("电池电压过低: %.1f V (最小: %.1f V)", batteryVoltage, m.config.Thresholds.Can.BatteryVoltage.Min),
				data)
		} else if batteryVoltage > m.config.Thresholds.Can.BatteryVoltage.Max {
			m.createAlert(data.VehicleID, data.VehicleType, "battery_voltage_high", "warning",
				fmt.Sprintf("电池电压过高: %.1f V (最大: %.1f V)", batteryVoltage, m.config.Thresholds.Can.BatteryVoltage.Max),
				data)
		}
	}

	// 检查燃油量
	if fuelLevel, ok := data.Signals["fuel_level"]; ok {
		if fuelLevel < m.config.Thresholds.Can.FuelLevel.Low {
			m.createAlert(data.VehicleID, data.VehicleType, "fuel_level_low", "info",
				fmt.Sprintf("燃油量低: %.1f%% (低燃油警告: %.1f%%)", fuelLevel, m.config.Thresholds.Can.FuelLevel.Low),
				data)
		}
	}
}

// checkObstacleThresholds 检查障碍物阈值
func (m *MonitorService) checkObstacleThresholds(data *models.TriggerData) {
	// 障碍物距离就是触发值
	distance := data.TriggerValue

	if distance < m.config.Thresholds.Perception.ObstacleDistance.Critical {
		m.createAlert(data.VehicleID, data.VehicleType, "obstacle_critical", "critical",
			fmt.Sprintf("障碍物距离临界: %.1f m (临界: %.1f m)", distance, m.config.Thresholds.Perception.ObstacleDistance.Critical),
			data)
	} else if distance < m.config.Thresholds.Perception.ObstacleDistance.Warning {
		m.createAlert(data.VehicleID, data.VehicleType, "obstacle_warning", "warning",
			fmt.Sprintf("障碍物距离警告: %.1f m (警告: %.1f m)", distance, m.config.Thresholds.Perception.ObstacleDistance.Warning),
			data)
	}
}

// checkLaneDeviationThresholds 检查车道偏离阈值
func (m *MonitorService) checkLaneDeviationThresholds(data *models.TriggerData) {
	// 车道偏离值就是触发值
	deviation := data.TriggerValue

	if deviation > m.config.Thresholds.Perception.LaneDeviation.Max {
		m.createAlert(data.VehicleID, data.VehicleType, "lane_deviation", "warning",
			fmt.Sprintf("车道偏离过大: %.2f m (最大: %.2f m)", deviation, m.config.Thresholds.Perception.LaneDeviation.Max),
			data)
	}
}

// createAlert 创建告警
func (m *MonitorService) createAlert(vehicleID, vehicleType, alertType, severity, message string, sourceData interface{}) {
	alert := &models.Alert{
		ID:          uuid.New().String(),
		Timestamp:   time.Now(),
		VehicleID:   vehicleID,
		VehicleType: vehicleType,
		AlertType:   alertType,
		Severity:    severity,
		Message:     message,
		SourceData:  sourceData,
	}

	// 将告警放入通道
	select {
	case m.alertChan <- alert:
		// 成功放入通道
	default:
		// 通道已满，直接处理
		log.Printf("警告: 告警通道已满，直接处理告警: %s\n", alert.Message)
		m.processAlert(alert)
	}
}

// processAlert 处理告警
func (m *MonitorService) processAlert(alert *models.Alert) {
	// 记录告警
	log.Printf("告警 [%s] [%s] [%s]: %s\n", alert.Severity, alert.VehicleID, alert.AlertType, alert.Message)
}

// ParseCANFile 解析CAN文件并处理数据
func (m *MonitorService) ParseCANFile(filePath, vehicleID, vehicleType string) error {
	// 使用解析服务解析CAN文件
	log.Printf("开始解析CAN文件: %s", filePath)
	canData, err := m.parserService.ParseCANFile(filePath, vehicleID, vehicleType)
	if err != nil {
		return fmt.Errorf("解析CAN文件失败: %v", err)
	}

	// 处理解析结果
	log.Printf("成功解析CAN文件，共%d条数据", len(canData))
	for _, data := range canData {
		// 将解析结果传递给监控服务处理
		if err := m.ProcessCanData(data); err != nil {
			log.Printf("处理CAN数据失败: %v", err)
		}
	}

	return nil
}