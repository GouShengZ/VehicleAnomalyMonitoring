package parser

import (
	"fmt"
	"log"

	"github.com/zhangyuchen/AutoDataHub-monitor/internal/models"
	"github.com/zhangyuchen/AutoDataHub-monitor/internal/parser/can"
	"github.com/zhangyuchen/AutoDataHub-monitor/internal/parser/dbc"
)

// ParserService 解析服务
type ParserService struct {
	dbcParser *dbc.DBCParser
	canParser *can.CANParser
	config    *models.Config
}

// NewParserService 创建新的解析服务
func NewParserService(config *models.Config) (*ParserService, error) {
	// 检查DBC配置
	if !config.DBC.Enabled {
		return nil, fmt.Errorf("DBC解析未启用")
	}

	// 创建DBC解析器
	dbcParser, err := dbc.NewDBCParser(config.DBC.FilePath)
	if err != nil {
		return nil, fmt.Errorf("创建DBC解析器失败: %v", err)
	}

	// 解析DBC文件
	if err := dbcParser.Parse(); err != nil {
		return nil, fmt.Errorf("解析DBC文件失败: %v", err)
	}

	// 创建CAN解析器
	canParser := can.NewCANParser(dbcParser)

	return &ParserService{
		dbcParser: dbcParser,
		canParser: canParser,
		config:    config,
	}, nil
}

// ParseCANFile 解析CAN文件
func (p *ParserService) ParseCANFile(filePath, vehicleID, vehicleType string) ([]*models.CanData, error) {
	log.Printf("开始解析CAN文件: %s, 车辆ID: %s, 车辆类型: %s", filePath, vehicleID, vehicleType)
	
	// 使用CAN解析器解析文件
	result, err := p.canParser.ParseCANFile(filePath, vehicleID, vehicleType)
	if err != nil {
		return nil, fmt.Errorf("解析CAN文件失败: %v", err)
	}

	log.Printf("成功解析CAN文件: %s, 共解析 %d 条数据", filePath, len(result))
	return result, nil
}

// ParseCANStream 解析CAN流数据
func (p *ParserService) ParseCANStream(data []byte, vehicleID, vehicleType string) (*models.CanData, error) {
	// 使用CAN解析器解析流数据
	return p.canParser.ParseCANStream(data, vehicleID, vehicleType)
}

// GetDBCMessages 获取所有DBC消息定义
func (p *ParserService) GetDBCMessages() map[uint32]*models.DBCMessage {
	return p.dbcParser.GetMessages()
}

// GetDBCMessage 根据ID获取DBC消息定义
func (p *ParserService) GetDBCMessage(id uint32) (*models.DBCMessage, bool) {
	return p.dbcParser.GetMessage(id)
}