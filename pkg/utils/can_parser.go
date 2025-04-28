package utils

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"go.einride.tech/can/pkg/dbc"
)

type CANParser struct {
	db         *dbc.File
	targetSigs map[string]struct{}
	lineNumber int
	timestamp  int64
	canID      uint32
	data       []byte
}

// NewCANParser 创建新的CAN解析器实例
func NewCANParser(db *dbc.File, targetSignals []string) *CANParser {
	targetSet := make(map[string]struct{}, len(targetSignals))
	for _, sig := range targetSignals {
		targetSet[sig] = struct{}{}
	}
	return &CANParser{
		db:         db,
		targetSigs: targetSet,
	}
}

// ParseLine 解析单行CAN日志
func (p *CANParser) ParseLine(line string) (timestamp int64, signals map[string]float64, err error) {
	p.lineNumber++
	line = strings.TrimSpace(line)
	if line == "" {
		return 0, nil, nil
	}

	// 重置解析状态
	p.timestamp = 0
	p.canID = 0
	p.data = nil
	signals = make(map[string]float64)

	if err := p.parseTimestamp(line); err != nil {
		return 0, nil, fmt.Errorf("行%d: %w", p.lineNumber, err)
	}
	if err := p.parseCANFrame(line); err != nil {
		return 0, nil, fmt.Errorf("行%d: %w", p.lineNumber, err)
	}
	signals, err = p.processCANMessage()
	if err != nil {
		return 0, nil, fmt.Errorf("行%d: %w", p.lineNumber, err)
	}
	return p.timestamp, signals, nil
}

func (p *CANParser) parseTimestamp(line string) error {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return fmt.Errorf("无效行格式，期望至少3个字段")
	}

	timestampStr := strings.Trim(parts[0], "()")
	ts, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("时间戳解析失败 '%s': %w", timestampStr, err)
	}
	p.timestamp = ts
	return nil
}

func (p *CANParser) parseCANFrame(line string) error {
	parts := strings.Fields(line)
	canFrame := parts[2]
	frameParts := strings.Split(canFrame, "#")
	if len(frameParts) != 2 {
		return fmt.Errorf("CAN帧格式无效 '%s'", canFrame)
	}

	idStr := frameParts[0]
	dataStr := frameParts[1]

	// 解析CAN ID
	id, err := strconv.ParseUint(idStr, 16, 32)
	if err != nil {
		return fmt.Errorf("CAN ID解析失败 '%s': %w", idStr, err)
	}
	p.canID = uint32(id)

	// 解码十六进制数据
	if len(dataStr)%2 != 0 {
		return fmt.Errorf("CAN数据长度必须为偶数，当前长度%d", len(dataStr))
	}
	data, err := hex.DecodeString(dataStr)
	if err != nil {
		return fmt.Errorf("CAN数据解码失败: %w", err)
	}
	p.data = data
	return nil
}

func (p *CANParser) processCANMessage() (map[string]float64, error) {
	signals := make(map[string]float64)
	var msgDef *dbc.MessageDef
	for _, def := range p.db.Defs {
		if m, ok := def.(*dbc.MessageDef); ok && uint32(m.MessageID) == p.canID {
			msgDef = m
			break
		}
	}
	if msgDef == nil {
		return nil, fmt.Errorf("CAN ID %X 未在DBC中定义", p.canID)
	}

	for i := range msgDef.Signals {
		sigDef := msgDef.Signals[i]
		if _, exists := p.targetSigs[string(sigDef.Name)]; !exists {
			continue
		}

		rawValue, err := parseSignalValue(&sigDef, p.data)
		if err != nil {
			return nil, fmt.Errorf("信号'%s'原始值解析失败: %w", sigDef.Name, err)
		}
		floatValue := float64(rawValue)*sigDef.Factor + sigDef.Offset

		signals[string(sigDef.Name)] = floatValue
	}

	return signals, nil
}

// convertPhysicalToFloat 将物理值安全转换为float64
func convertPhysicalToFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("不支持的数值类型: %T", value)
	}
}

// parseSignalValue 从CAN数据中提取信号的原始整数值
func parseSignalValue(sigDef *dbc.SignalDef, data []byte) (int64, error) {
	startBit := int(sigDef.StartBit)
	length := int(sigDef.Size)

	if startBit+length > len(data)*8 {
		return 0, fmt.Errorf("信号 %s 超出数据范围", sigDef.Name)
	}

	var value int64
	for i := 0; i < length; i++ {
		bitPos := startBit + i
		bytePos := bitPos / 8
		bitOffset := uint(7 - (bitPos % 8))
		bitValue := (int64(data[bytePos]) >> bitOffset) & 0x01
		value |= bitValue << (length - i - 1)
	}

	// 处理符号扩展
	if sigDef.IsSigned {
		if value&(1<<(length-1)) != 0 {
			value |= ^((1 << length) - 1)
		}
	}

	return value, nil
}
