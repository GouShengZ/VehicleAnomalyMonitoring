package models

import (
	"time"
)

// Datalog 表示数据日志的结构体
type DataLogs struct {
	ID                int       `gorm:"column:id;type:smallint(6);primary_key;AUTO_INCREMENT" json:"id"`
	UpdatedAt         time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreateAt          time.Time `gorm:"column:create_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"create_at"`
	Vin               string    `gorm:"column:vin;type:varchar(17);NOT NULL" json:"vin"`
	TriggerTimestamp  int64     `gorm:"column:trigger_timestamp;type:timestamp;NOT NULL" json:"trigger_timestamp"`
	CarType           string    `gorm:"column:car_type;type:varchar(255);NOT NULL" json:"car_type"`
	UseType           string    `gorm:"column:use_type;type:varchar(255);NOT NULL" json:"use_type"`
	TriggerID         string    `gorm:"column:trigger_id;type:varchar(255);NOT NULL" json:"trigger_id"`
	IsCrash           int       `gorm:"column:is_crash;type:int(11);NOT NULL" json:"is_crash"`
	CrashReason       string    `gorm:"column:crash_reason;type:varchar(2000);NOT NULL" json:"crash_reason"`
	CriterionJudgment string    `gorm:"column:criterion_judgment;type:varchar(2000);NOT NULL" json:"criterion_judgment"`
}

func (m *DataLogs) TableName() string {
	return "data_logs"
}
