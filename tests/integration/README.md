# 集成测试文档## 概述本目录包含AutoDataHub-monitor项目的集成测试，用于验证系统与外部服务（如MySQL、Redis）的集成。## 测试结构```tests/integration/├── README.md           # 本文档├── mysql_test.go       # MySQL集成测试└── redis_test.go       # Redis集成测试```## 环境要求### Docker环境集成测试需要Docker和Docker Compose来启动测试环境中的服务。### 环境变量测试使用以下环境变量配置：#### MySQL配置- `MYSQL_HOST` - MySQL主机地址（默认: localhost）- `MYSQL_PORT` - MySQL端口（默认: 3306）- `MYSQL_USER` - MySQL用户名（默认: root）- `MYSQL_PASSWORD` - MySQL密码（默认: password）- `MYSQL_DATABASE` - MySQL数据库名（默认: testdb）#### Redis配置- `REDIS_HOST` - Redis主机地址（默认: localhost）- `REDIS_PORT` - Redis端口（默认: 6379）- `REDIS_PASSWORD` - Redis密码（默认: 空）## 运行测试### 1. 使用Docker Compose（推荐）```bash# 启动测试环境make test-integration# 或者手动执行docker-compose -f docker-compose.test.yml up -dsleep 10  # 等待服务启动go test -v ./tests/integration/...docker-compose -f docker-compose.test.yml down```### 2. 使用本地服务如果你已经在本地运行了MySQL和Redis，可以直接运行测试：```bashgo test -v ./tests/integration/...```### 3. 跳过集成测试运行单元测试时跳过集成测试：```bashgo test -short ./...```## 测试内容### MySQL集成测试 (`mysql_test.go`)测试以下功能：- **数据库连接测试** - 验证数据库连接和ping操作- **表操作测试** - 创建、删除表- **数据操作测试** - 插入、查询、更新数据- **事务测试** - 事务提交和回滚- **连接池测试** - 并发连接测试### Redis集成测试 (`redis_test.go`)测试以下功能：- **基本连接测试** - 验证Redis连接和ping操作- **字符串操作** - SET、GET、TTL操作- **哈希操作** - HSET、HGET、HGETALL操作- **列表操作** - LPUSH、LPOP、LRANGE操作- **集合操作** - SADD、SISMEMBER、SMEMBERS操作- **有序集合操作** - ZADD、ZRANGE、ZSCORE操作- **发布订阅测试** - PUBLISH、SUBSCRIBE操作- **连接池测试** - 并发操作和管道操作## 测试配置### Docker测试环境配置测试环境在`docker-compose.test.yml`中定义：```yamlservices:  mysql:    image: mysql:8.0    environment:      MYSQL_ROOT_PASSWORD: password      MYSQL_DATABASE: testdb    ports:      - "3306:3306"      redis:    image: redis:7-alpine    ports:      - "6379:6379"```### 超时和重试- **连接重试**: 测试会重试连接最多30次，每次间隔1秒- **测试超时**: 每个测试用例有合理的超时设置- **发布订阅超时**: PubSub测试使用5秒超时## 故障排查### 常见问题1. **连接拒绝错误**   ```   dial tcp 127.0.0.1:3306: connect: connection refused   ```   **解决方案**: 确保MySQL/Redis服务正在运行2. **认证失败**   ```   Error 1045: Access denied for user 'root'@'localhost'   ```   **解决方案**: 检查用户名和密码配置3. **数据库不存在**   ```   Error 1049: Unknown database 'testdb'   ```   **解决方案**: 确保测试数据库已创建### 调试技巧1. **查看服务日志**:   ```bash   docker-compose -f docker-compose.test.yml logs mysql   docker-compose -f docker-compose.test.yml logs redis   ```2. **手动连接测试**:   ```bash   # MySQL   docker exec -it <mysql_container> mysql -uroot -ppassword testdb      # Redis   docker exec -it <redis_container> redis-cli   ```3. **运行单个测试**:   ```bash   go test -v ./tests/integration/ -run TestMySQLIntegration   go test -v ./tests/integration/ -run TestRedisIntegration
   ```

## 最佳实践

1. **测试隔离**: 每个测试用例都会清理自己创建的数据
2. **并发安全**: 测试使用唯一的键名避免冲突
3. **资源清理**: 使用defer语句确保资源正确释放
4. **错误处理**: 明确的错误消息便于调试
5. **超时控制**: 适当的超时避免测试挂起

## 持续集成

集成测试已集成到CI/CD流水线中（`.github/workflows/ci.yml`），会在以下情况下运行：
- 推送到main分支
- 创建Pull Request
- 手动触发

CI环境使用Docker服务容器提供MySQL和Redis服务。
