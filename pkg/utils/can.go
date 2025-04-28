package utils

import (
	"bufio"
	"fmt"
	"os"
	"sort"

	"go.einride.tech/can/pkg/dbc"
)

// ParseCANLogWithDBC 使用 DBC 文件解析 CAN 日志文件，提取指定信号的值。
// 参数:
//
//	canLogPath: CAN 日志文件的路径 (期望为 CANDump 格式)。
//	dbcPath: DBC 文件的路径。
//	targetSignals: 需要解析的信号名称列表。
//
// 返回:
//
//	signalData: 一个 map，第一层 key 是时间戳 (int64)，value 是另一个 map。
//	            内层 map 的 key 是信号名 (string)，value 是信号值 (float64)。
//	timestamps: 一个排序后的时间戳数组 (int64)，用于顺序访问数据。
//	error: 解析过程中发生的任何错误。
func ParseCANLogWithDBC(canLogPath, dbcPath string, targetSignals []string) (map[int64]map[string]float64, []int64, error) {
	// 1. 解析 DBC 文件
	dbcFileContent, err := os.ReadFile(dbcPath)
	if err != nil {
		return nil, nil, fmt.Errorf("读取 DBC 文件 '%s' 失败: %w", dbcPath, err)
	}
	// 创建DBC解析器并解析文件内容
	dbcParser := dbc.NewParser(dbcPath, dbcFileContent)
	if err := dbcParser.Parse(); err != nil {
		return nil, nil, fmt.Errorf("解析 DBC 文件 '%s' 失败: %w", dbcPath, err)
	}
	db := dbcParser.File()

	// 初始化CAN解析器
	canParser := NewCANParser(db, targetSignals)

	// 3. 打开并读取 CAN 日志文件
	canFile, err := os.Open(canLogPath)
	if err != nil {
		return nil, nil, fmt.Errorf("打开 CAN 日志文件 '%s' 失败: %w", canLogPath, err)
	}
	defer canFile.Close()

	signalData := make(map[int64]map[string]float64)
	timestampSet := make(map[int64]struct{})

	scanner := bufio.NewScanner(canFile)
	for scanner.Scan() {
		timestamp, signals, err := canParser.ParseLine(scanner.Text())
		if err != nil {
			return nil, nil, fmt.Errorf("解析错误[%s:%d]: %w", canLogPath, canParser.lineNumber, err)
		}
		if timestamp == 0 || len(signals) == 0 {
			continue
		}

		// 收集信号数据
		if _, exists := signalData[timestamp]; !exists {
			signalData[timestamp] = make(map[string]float64)
		}
		for sigName, value := range signals {
			signalData[timestamp][sigName] = value
		}
		timestampSet[timestamp] = struct{}{}

	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("读取 CAN 日志文件 '%s' 时出错: %w", canLogPath, err)
	}

	// 6. 整理并排序时间戳
	timestamps := make([]int64, 0, len(timestampSet))
	for ts := range timestampSet {
		timestamps = append(timestamps, ts)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })

	if len(signalData) == 0 {
		return nil, nil, fmt.Errorf("未找到有效信号数据，请检查DBC匹配和日志格式")
	}
	return signalData, timestamps, nil
}
