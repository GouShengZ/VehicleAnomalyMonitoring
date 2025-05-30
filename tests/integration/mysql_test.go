package integration

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQLIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 从环境变量获取数据库连接信息
	host := getEnvOrDefault("MYSQL_HOST", "localhost")
	port := getEnvOrDefault("MYSQL_PORT", "3306")
	user := getEnvOrDefault("MYSQL_USER", "root")
	password := getEnvOrDefault("MYSQL_PASSWORD", "password")
	database := getEnvOrDefault("MYSQL_DATABASE", "testdb")

	// 构建数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, database)

	// 等待数据库启动
	var db *sql.DB
	var err error

	// 重试连接数据库
	for i := 0; i < 30; i++ {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}

	require.NoError(t, err, "无法连接到MySQL数据库")
	defer db.Close()

	t.Run("数据库连接测试", func(t *testing.T) {
		err := db.Ping()
		assert.NoError(t, err, "数据库ping失败")
	})

	t.Run("创建表测试", func(t *testing.T) {
		createTableSQL := `CREATE TABLE IF NOT EXISTS test_vehicles (
			id INT AUTO_INCREMENT PRIMARY KEY,
			vehicle_id VARCHAR(50) NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			speed FLOAT,
			location_lat DECIMAL(10, 8),
			location_lng DECIMAL(11, 8),
			status VARCHAR(20),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`

		_, err := db.Exec(createTableSQL)
		assert.NoError(t, err, "创建表失败")

		// 清理
		defer func() {
			_, _ = db.Exec("DROP TABLE IF EXISTS test_vehicles")
		}()
	})

	t.Run("插入和查询数据测试", func(t *testing.T) {
		// 先创建表
		createTableSQL := `CREATE TABLE IF NOT EXISTS test_data (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100),
			value INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`

		_, err := db.Exec(createTableSQL)
		require.NoError(t, err)

		// 插入测试数据
		insertSQL := "INSERT INTO test_data (name, value) VALUES (?, ?)"
		result, err := db.Exec(insertSQL, "test_name", 42)
		assert.NoError(t, err, "插入数据失败")

		// 检查插入结果
		lastID, err := result.LastInsertId()
		assert.NoError(t, err)
		assert.Greater(t, lastID, int64(0))

		// 查询数据
		var name string
		var value int
		selectSQL := "SELECT name, value FROM test_data WHERE id = ?"
		err = db.QueryRow(selectSQL, lastID).Scan(&name, &value)
		assert.NoError(t, err, "查询数据失败")
		assert.Equal(t, "test_name", name)
		assert.Equal(t, 42, value)

		// 清理
		defer func() {
			_, _ = db.Exec("DROP TABLE IF EXISTS test_data")
		}()
	})

	t.Run("事务测试", func(t *testing.T) {
		// 创建测试表
		createTableSQL := `CREATE TABLE IF NOT EXISTS test_transaction (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100)
		)`

		_, err := db.Exec(createTableSQL)
		require.NoError(t, err)

		// 开始事务
		tx, err := db.Begin()
		require.NoError(t, err)

		// 在事务中插入数据
		_, err = tx.Exec("INSERT INTO test_transaction (name) VALUES (?)", "transaction_test")
		assert.NoError(t, err)

		// 回滚事务
		err = tx.Rollback()
		assert.NoError(t, err)

		// 验证数据未被提交
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM test_transaction WHERE name = ?", "transaction_test").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "事务回滚后数据应该不存在")

		// 清理
		defer func() {
			_, _ = db.Exec("DROP TABLE IF EXISTS test_transaction")
		}()
	})
}

func TestMySQLConnectionPool(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	host := getEnvOrDefault("MYSQL_HOST", "localhost")
	port := getEnvOrDefault("MYSQL_PORT", "3306")
	user := getEnvOrDefault("MYSQL_USER", "root")
	password := getEnvOrDefault("MYSQL_PASSWORD", "password")
	database := getEnvOrDefault("MYSQL_DATABASE", "testdb")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, database)

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer db.Close()

	// 配置连接池
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// 测试并发连接
	t.Run("并发连接测试", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()

				err := db.Ping()
				assert.NoError(t, err, fmt.Sprintf("连接%d ping失败", id))
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// getEnvOrDefault 获取环境变量或返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
