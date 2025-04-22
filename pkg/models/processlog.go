package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/zhangyuchen/AutoDataHub-monitor/configs"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProcessLog 表示处理日志的结构体
type ProcessLogs struct {
	ID               int       `gorm:"column:id;type:smallint(6);primary_key;AUTO_INCREMENT" json:"id"`
	UpdatedAt        time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreateAt         time.Time `gorm:"column:create_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"create_at"`
	Vin              string    `gorm:"column:vin;type:varchar(17);NOT NULL" json:"vin"`
	TriggerTimestamp int64     `gorm:"column:trigger_timestamp;type:timestamp;NOT NULL" json:"trigger_timestamp"`
	CarType          string    `gorm:"column:car_type;type:varchar(255);NOT NULL" json:"car_type"`
	UseType          string    `gorm:"column:use_type;type:varchar(255);NOT NULL" json:"use_type"`
	TriggerID        string    `gorm:"column:trigger_id;type:varchar(255);NOT NULL" json:"trigger_id"`
	ProcessStatus    string    `gorm:"column:process_status;type:varchar(50);NOT NULL" json:"process_status"`
	ProcessLog       string    `gorm:"column:process_log;type:varchar(2000);NOT NULL" json:"process_log"`
}

func (m *ProcessLogs) TableName() string {
	return "process_logs"
}

func CreateProcessLog(db *gorm.DB, log ProcessLogs) (*ProcessLogs, error) {
	result := db.Table("process_logs").Create(&log)
	if result.Error != nil {
		configs.Client.Logger.Error("failed to create process log", zap.Error(result.Error))
		return nil, fmt.Errorf("failed to create process log: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		configs.Client.Logger.Error("no rows affected")
		return nil, errors.New("no rows affected")
	}
	return &log, nil
}

func UpdateProcessLog(db *gorm.DB, data map[string]interface{}) error {
	result := db.Model(&ProcessLogs{}).
		Where("id = ?", data["id"]).
		Updates(data)
	if result.Error != nil {
		configs.Client.Logger.Error("failed to update process log",
			zap.Any("data", data),
			zap.Error(result.Error))
		return fmt.Errorf("failed to update process log: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		configs.Client.Logger.Warn("no rows updated",
			zap.Any("data", data))
		return errors.New("no rows updated")
	}
	return nil
}

func AddProcessLog(db *gorm.DB, logId int, queueName string) error {
	result := db.Table("process_logs").
		Where("id = ?", logId).
		Update("process_status", gorm.Expr("CONCAT(process_log, ?)", " -> "+queueName))
	if result.Error != nil {
		configs.Client.Logger.Error("failed to add process log", zap.Error(result.Error))
		return fmt.Errorf("failed to add process log: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		configs.Client.Logger.Warn("no rows updated")
		return errors.New("no rows updated")
	}
	return nil
}
