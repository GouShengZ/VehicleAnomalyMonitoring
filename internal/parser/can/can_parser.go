package can

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/zhangyuchen/AutoDataHub-monitor/internal/models"
	"github.com/zhangyuchen/AutoDataHub-monitor/internal/parser/dbc"
)

// CANFrame CAN帧结构
type CANFrame struct {
	Timestamp time.Time // 时间戳
	ID        uint32    // CAN ID
	Data      []byte    // 数据
	Length    int       // 数据长度
}

// CANParser CAN文件解析器
type CANParser struct {
	dbcParser *dbc.DBCParser
}

// NewCANParser 创建新的CAN解析器
func NewCANParser(dbcParser *dbc.DBCParser) *CANParser {
	return &CANParser{
		dbcParser: dbcParser,
	}
}

// ParseCANFile 解析CAN文件
func (p *CANParser) ParseCANFile(filePath string, vehicleID, vehicleType string) ([]*models.CanData, error) {
	// 打开CAN文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开CAN文件失败: %v", err)
	}
	defer file.Close()

	// 解析CAN文件
	frames, err := p.readCANFrames(file)
	if err != nil {
		return nil, fmt.Errorf("读取CAN帧失败: %v", err)
	}

	// 将CAN帧转换为CanData
	result := make([]*models.CanData, 0, len(frames))
	for _, frame := range frames {
		// 使用DBC解析器解码CAN帧
		signals, err := p.dbcParser.DecodeCANFrame(frame.ID, frame.Data)
		if err != nil {
			// 记录错误但继续处理
			fmt.Printf("解码CAN帧失败 (ID: %d): %v\n", frame.ID, err)
			continue
		}

		// 创建CanData
		canData := &models.CanData{
			Timestamp:   frame.Timestamp,
			VehicleID:   vehicleID,
			VehicleType: vehicleType,
			Signals:     signals,
			RawData:     frame.Data,
		}

		result = append(result, canData)
	}

	return result, nil
}

// readCANFrames 从文件中读取CAN帧
// 注意：这里假设一种简单的CAN文件格式，实际应用中可能需要根据具体格式调整
func (p *CANParser) readCANFrames(reader io.Reader) ([]*CANFrame, error) {
	var frames []*CANFrame

	// 读取文件头部信息（如果有）
	// ...

	// 读取CAN帧
	buffer := make([]byte, 16) // 假设每个CAN帧记录为16字节
	for {
		// 读取一个CAN帧记录
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if n < 16 {
			continue // 数据不完整，跳过
		}

		// 解析时间戳（假设前8字节是时间戳，单位为微秒）
		timestampMicros := binary.LittleEndian.Uint64(buffer[0:8])
		timestamp := time.Unix(0, int64(timestampMicros)*1000) // 转换为纳秒

		// 解析CAN ID（假设接下来的4字节是CAN ID）
		id := binary.LittleEndian.Uint32(buffer[8:12])

		// 解析数据长度（假设接下来的1字节是数据长度）
		length := int(buffer[12])
		if length > 8 {
			length = 8 // CAN数据最大8字节
		}

		// 解析数据（假设接下来的8字节是数据）
		data := make([]byte, length)
		copy(data, buffer[13:13+length])

		// 创建CAN帧
		frame := &CANFrame{
			Timestamp: timestamp,
			ID:        id,
			Data:      data,
			Length:    length,
		}

		frames = append(frames, frame)
	}

	return frames, nil
}

// ParseCANStream 解析CAN流数据
func (p *CANParser) ParseCANStream(data []byte, vehicleID, vehicleType string) (*models.CanData, error) {
	// 确保数据长度足够
	if len(data) < 13 { // 至少需要时间戳(8) + ID(4) + 长度(1)
		return nil, fmt.Errorf("CAN数据长度不足")
	}

	// 解析时间戳（假设前8字节是时间戳，单位为微秒）
	timestampMicros := binary.LittleEndian.Uint64(data[0:8])
	timestamp := time.Unix(0, int64(timestampMicros)*1000) // 转换为纳秒

	// 解析CAN ID（假设接下来的4字节是CAN ID）
	id := binary.LittleEndian.Uint32(data[8:12])

	// 解析数据长度（假设接下来的1字节是数据长度）
	length := int(data[12])
	if length > 8 {
		length = 8 // CAN数据最大8字节
	}

	// 确保数据长度足够
	if len(data) < 13+length {
		return nil, fmt.Errorf("CAN数据长度不足")
	}

	// 解析数据
	canData := make([]byte, length)
	copy(canData, data[13:13+length])

	// 使用DBC解析器解码CAN帧
	signals, err := p.dbcParser.DecodeCANFrame(id, canData)
	if err != nil {
		return nil, fmt.Errorf("解码CAN帧失败: %v", err)
	}

	// 创建CanData
	result := &models.CanData{
		Timestamp:   timestamp,
		VehicleID:   vehicleID,
		VehicleType: vehicleType,
		Signals:     signals,
		RawData:     canData,
	}

	return result, nil
}