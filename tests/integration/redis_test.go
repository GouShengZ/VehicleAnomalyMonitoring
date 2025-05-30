package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 从环境变量获取Redis连接信息
	host := getEnvOrDefault("REDIS_HOST", "localhost")
	port := getEnvOrDefault("REDIS_PORT", "6379")
	password := getEnvOrDefault("REDIS_PASSWORD", "")

	// 创建Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0, // 使用默认数据库
	})
	defer rdb.Close()

	ctx := context.Background()

	// 等待Redis启动
	var err error
	for i := 0; i < 30; i++ {
		_, err = rdb.Ping(ctx).Result()
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	require.NoError(t, err, "无法连接到Redis")

	t.Run("基本连接测试", func(t *testing.T) {
		pong, err := rdb.Ping(ctx).Result()
		assert.NoError(t, err, "Redis ping失败")
		assert.Equal(t, "PONG", pong)
	})

	t.Run("字符串操作测试", func(t *testing.T) {
		key := "test:string"
		value := "test_value"

		// 设置值
		err := rdb.Set(ctx, key, value, time.Minute).Err()
		assert.NoError(t, err, "设置字符串值失败")

		// 获取值
		result, err := rdb.Get(ctx, key).Result()
		assert.NoError(t, err, "获取字符串值失败")
		assert.Equal(t, value, result)

		// 检查TTL
		ttl, err := rdb.TTL(ctx, key).Result()
		assert.NoError(t, err, "获取TTL失败")
		assert.Greater(t, ttl, time.Duration(0))

		// 清理
		defer rdb.Del(ctx, key)
	})

	t.Run("哈希操作测试", func(t *testing.T) {
		key := "test:hash"
		field := "field1"
		value := "value1"

		// 设置哈希字段
		err := rdb.HSet(ctx, key, field, value).Err()
		assert.NoError(t, err, "设置哈希字段失败")

		// 获取哈希字段
		result, err := rdb.HGet(ctx, key, field).Result()
		assert.NoError(t, err, "获取哈希字段失败")
		assert.Equal(t, value, result)

		// 获取所有字段
		all, err := rdb.HGetAll(ctx, key).Result()
		assert.NoError(t, err, "获取所有哈希字段失败")
		assert.Equal(t, value, all[field])

		// 清理
		defer rdb.Del(ctx, key)
	})

	t.Run("列表操作测试", func(t *testing.T) {
		key := "test:list"
		values := []string{"item1", "item2", "item3"}

		// 推入列表
		for _, v := range values {
			err := rdb.LPush(ctx, key, v).Err()
			assert.NoError(t, err, "推入列表失败")
		}

		// 获取列表长度
		length, err := rdb.LLen(ctx, key).Result()
		assert.NoError(t, err, "获取列表长度失败")
		assert.Equal(t, int64(len(values)), length)

		// 获取列表元素
		result, err := rdb.LRange(ctx, key, 0, -1).Result()
		assert.NoError(t, err, "获取列表元素失败")
		assert.Equal(t, len(values), len(result))

		// 弹出元素
		popped, err := rdb.LPop(ctx, key).Result()
		assert.NoError(t, err, "弹出列表元素失败")
		assert.Equal(t, values[len(values)-1], popped) // LPUSH是从左边推入，所以最后一个元素先被弹出

		// 清理
		defer rdb.Del(ctx, key)
	})

	t.Run("集合操作测试", func(t *testing.T) {
		key := "test:set"
		members := []string{"member1", "member2", "member3"}

		// 添加集合成员
		for _, m := range members {
			err := rdb.SAdd(ctx, key, m).Err()
			assert.NoError(t, err, "添加集合成员失败")
		}

		// 检查成员是否存在
		exists, err := rdb.SIsMember(ctx, key, members[0]).Result()
		assert.NoError(t, err, "检查集合成员失败")
		assert.True(t, exists)

		// 获取所有成员
		allMembers, err := rdb.SMembers(ctx, key).Result()
		assert.NoError(t, err, "获取集合成员失败")
		assert.Equal(t, len(members), len(allMembers))

		// 清理
		defer rdb.Del(ctx, key)
	})

	t.Run("有序集合操作测试", func(t *testing.T) {
		key := "test:zset"

		// 添加有序集合成员
		members := []*redis.Z{
			{Score: 1, Member: "one"},
			{Score: 2, Member: "two"},
			{Score: 3, Member: "three"},
		}

		err := rdb.ZAdd(ctx, key, members...).Err()
		assert.NoError(t, err, "添加有序集合成员失败")

		// 获取指定范围的成员
		result, err := rdb.ZRange(ctx, key, 0, -1).Result()
		assert.NoError(t, err, "获取有序集合成员失败")
		assert.Equal(t, 3, len(result))
		assert.Equal(t, "one", result[0])

		// 获取成员分数
		score, err := rdb.ZScore(ctx, key, "two").Result()
		assert.NoError(t, err, "获取成员分数失败")
		assert.Equal(t, float64(2), score)

		// 清理
		defer rdb.Del(ctx, key)
	})

	t.Run("发布订阅测试", func(t *testing.T) {
		channel := "test:channel"
		message := "test_message"

		// 创建订阅
		pubsub := rdb.Subscribe(ctx, channel)
		defer pubsub.Close()

		// 等待订阅确认
		_, err := pubsub.Receive(ctx)
		assert.NoError(t, err, "订阅失败")

		// 发布消息
		err = rdb.Publish(ctx, channel, message).Err()
		assert.NoError(t, err, "发布消息失败")

		// 接收消息（设置超时）
		msgCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		msg, err := pubsub.ReceiveMessage(msgCtx)
		assert.NoError(t, err, "接收消息失败")
		assert.Equal(t, channel, msg.Channel)
		assert.Equal(t, message, msg.Payload)
	})
}

func TestRedisConnectionPool(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	host := getEnvOrDefault("REDIS_HOST", "localhost")
	port := getEnvOrDefault("REDIS_PORT", "6379")
	password := getEnvOrDefault("REDIS_PASSWORD", "")

	// 创建Redis客户端with连接池配置
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Password:     password,
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		PoolTimeout:  30 * time.Second,
	})
	defer rdb.Close()

	ctx := context.Background()

	t.Run("并发连接测试", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()

				key := fmt.Sprintf("test:concurrent:%d", id)
				value := fmt.Sprintf("value%d", id)

				// 并发执行Redis操作
				err := rdb.Set(ctx, key, value, time.Minute).Err()
				assert.NoError(t, err, fmt.Sprintf("并发设置值失败: %d", id))

				result, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err, fmt.Sprintf("并发获取值失败: %d", id))
				assert.Equal(t, value, result)

				// 清理
				rdb.Del(ctx, key)
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("管道操作测试", func(t *testing.T) {
		pipe := rdb.Pipeline()

		// 批量操作
		keys := make([]string, 5)
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("test:pipeline:%d", i)
			keys[i] = key
			pipe.Set(ctx, key, fmt.Sprintf("value%d", i), time.Minute)
		}

		// 执行管道
		_, err := pipe.Exec(ctx)
		assert.NoError(t, err, "管道执行失败")

		// 验证结果
		for i, key := range keys {
			result, err := rdb.Get(ctx, key).Result()
			assert.NoError(t, err, "获取管道设置的值失败")
			assert.Equal(t, fmt.Sprintf("value%d", i), result)
		}

		// 清理
		for _, key := range keys {
			rdb.Del(ctx, key)
		}
	})
}
