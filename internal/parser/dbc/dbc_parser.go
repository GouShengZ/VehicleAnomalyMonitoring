package dbc

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/zhangyuchen/AutoDataHub-monitor/internal/models"
)

// DBCParser DBC文件解析器
type DBCParser struct {
	filePath string
	messages map[uint32]*models.DBCMessage
}

// NewDBCParser 创建新的DBC解析器
func NewDBCParser(filePath string) (*DBCParser, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("DBC文件不存在: %s", filePath)
	}

	return &DBCParser{
		filePath: filePath,
		messages: make(map[uint32]*models.DBCMessage),
	}, nil
}

// Parse 解析DBC文件
func (p *DBCParser) Parse() error {
	file, err := os.Open(p.filePath)
	if err != nil {
		return fmt.Errorf("打开DBC文件失败: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentMessage *models.DBCMessage

	// 正则表达式用于匹配不同类型的行
	messageRegex := regexp.MustCompile(`^BO_ (\d+) ([A-Za-z0-9_]+): (\d+) ([A-Za-z0-9_]+)`)
	signalRegex := regexp.MustCompile(`^SG_ ([A-Za-z0-9_]+) : (\d+)\|(\d+)@(\d+)(\+|-) \(([0-9.]+),([0-9.-]+)\) \[([0-9.-]+)\|([0-9.-]+)\] "([^"]*)" ([A-Za-z0-9_]+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// 匹配消息定义行
		if matches := messageRegex.FindStringSubmatch(line); matches != nil {
			id, _ := strconv.ParseUint(matches[1], 10, 32)
			dlc, _ := strconv.Atoi(matches[3])

			currentMessage = &models.DBCMessage{
				ID:      uint32(id),
				Name:    matches[2],
				DLC:     dlc,
				Signals: make(map[string]models.DBCSignal),
			}
			p.messages[uint32(id)] = currentMessage
		}

		// 匹配信号定义行
		if currentMessage != nil {
			if matches := signalRegex.FindStringSubmatch(line); matches != nil {
				startBit, _ := strconv.Atoi(matches[2])
				length, _ := strconv.Atoi(matches[3])
				factor, _ := strconv.ParseFloat(matches[6], 64)
				offset, _ := strconv.ParseFloat(matches[7], 64)
				minimum, _ := strconv.ParseFloat(matches[8], 64)
				maximum, _ := strconv.ParseFloat(matches[9], 64)

				isInteger := false
				if matches[4] == "1" {
					isInteger = true
				}

				signal := models.DBCSignal{
					Name:      matches[1],
					StartBit:  startBit,
					Length:    length,
					Factor:    factor,
					Offset:    offset,
					Minimum:   minimum,
					Maximum:   maximum,
					Unit:      matches[10],
					IsInteger: isInteger,
				}

				currentMessage.Signals[signal.Name] = signal
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取DBC文件失败: %v", err)
	}

	if len(p.messages) == 0 {
		return errors.New("DBC文件中未找到有效的消息定义")
	}

	return nil
}

// GetMessages 获取所有消息定义
func (p *DBCParser) GetMessages() map[uint32]*models.DBCMessage {
	return p.messages
}

// GetMessage 根据ID获取消息定义
func (p *DBCParser) GetMessage(id uint32) (*models.DBCMessage, bool) {
	message, ok := p.messages[id]
	return message, ok
}

// DecodeCANFrame 解码CAN帧数据
func (p *DBCParser) DecodeCANFrame(id uint32, data []byte) (map[string]float64, error) {
	message, ok := p.messages[id]
	if !ok {
		return nil, fmt.Errorf("未找到ID为%d的消息定义", id)
	}

	if len(data) < message.DLC {
		return nil, fmt.Errorf("数据长度不足: 期望%d字节，实际%d字节", message.DLC, len(data))
	}

	result := make(map[string]float64)

	for name, signal := range message.Signals {
		// 提取原始值
		rawValue := extractRawValue(data, signal.StartBit, signal.Length)

		// 应用因子和偏移
		value := float64(rawValue)*signal.Factor + signal.Offset

		// 存储解码后的值
		result[name] = value
	}

	return result, nil
}

// extractRawValue 从字节数组中提取原始值
func extractRawValue(data []byte, startBit, length int) uint64 {
	var result uint64

	// 计算起始字节和位偏移
	startByte := startBit / 8
	bitOffset := startBit % 8

	// 计算需要处理的字节数
	bytesNeeded := (bitOffset + length + 7) / 8

	// 确保不超出数据范围
	if startByte+bytesNeeded > len(data) {
		bytesNeeded = len(data) - startByte
	}

	// 从字节数组中提取值
	for i := 0; i < bytesNeeded; i++ {
		if startByte+i < len(data) {
			result |= uint64(data[startByte+i]) << (i * 8)
		}
	}

	// 应用位掩码
	result = (result >> bitOffset) & ((1 << length) - 1)

	return result
}
