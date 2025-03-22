package configs

import (
	"fmt"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLConfig 表示MySQL的配置信息
type MySQLConfig struct {
	Host     string `yaml:"host"`     // MySQL主机地址
	Port     int    `yaml:"port"`     // MySQL端口
	Username string `yaml:"username"` // MySQL用户名
	Password string `yaml:"password"` // MySQL密码
	Database string `yaml:"database"` // MySQL数据库名
	Charset  string `yaml:"charset"`  // 字符集
}

// 全局MySQL客户端实例
var (
	mysqlDB     *gorm.DB
	mysqlDBOnce sync.Once
)

// DefaultMySQLConfig 返回默认的MySQL配置
func DefaultMySQLConfig() *MySQLConfig {
	return &MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "",
		Database: "autodatahub",
		Charset:  "utf8mb4",
	}
}

// DSN 返回MySQL连接字符串
func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset)
}

// GetMySQLDB 获取全局MySQL客户端实例
func GetMySQLDB() *gorm.DB {
	mysqlDBOnce.Do(func() {
		// 确保配置已加载
		mysqlConfig := GetConfig().MySQL

		// 创建MySQL客户端
		db, err := gorm.Open(mysql.Open(mysqlConfig.DSN()), &gorm.Config{})
		if err != nil {
			fmt.Printf("连接MySQL数据库失败: %v\n", err)
			return
		}

		mysqlDB = db
	})

	return mysqlDB
}

// CloseMySQLDB 关闭MySQL客户端连接
func CloseMySQLDB() error {
	if mysqlDB != nil {
		sqlDB, err := mysqlDB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
