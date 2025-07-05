package com.vehicle.anomaly.sink;

import com.vehicle.anomaly.config.ConfigManager;
import com.vehicle.anomaly.model.AnomalyData;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.sink.RichSinkFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import redis.clients.jedis.Jedis;
import redis.clients.jedis.JedisPool;
import redis.clients.jedis.JedisPoolConfig;

import java.sql.*;
import java.time.format.DateTimeFormatter;

/**
 * 异常数据输出Sink - 优先写入Redis，失败时写入MySQL
 */
public class AnomalyDataSink extends RichSinkFunction<AnomalyData> {
    
    private static final Logger logger = LoggerFactory.getLogger(AnomalyDataSink.class);
    
    private transient JedisPool jedisPool;
    private transient Connection mysqlConnection;
    private transient ObjectMapper objectMapper;
    
    private String redisHost;
    private int redisPort;
    private String redisPassword;
    private int redisDatabase;
    private int redisTimeout;
    private String redisQueueKey;
    
    private String mysqlUrl;
    private String mysqlUsername;
    private String mysqlPassword;
    private String mysqlTable;
    
    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        
        // 初始化配置
        initializeConfig();
        
        // 初始化ObjectMapper
        objectMapper = new ObjectMapper();
        objectMapper.registerModule(new JavaTimeModule());
        
        // 初始化Redis连接池
        initializeRedis();
        
        // 初始化MySQL连接
        initializeMysql();
        
        logger.info("AnomalyDataSink initialized successfully");
    }
    
    private void initializeConfig() {
        redisHost = ConfigManager.getRedisHost();
        redisPort = ConfigManager.getRedisPort();
        redisPassword = ConfigManager.getRedisPassword();
        redisDatabase = ConfigManager.getRedisDatabase();
        redisTimeout = ConfigManager.getRedisTimeout();
        redisQueueKey = ConfigManager.getRedisQueueKey();
        
        mysqlUrl = String.format("jdbc:mysql://%s:%d/%s?useSSL=false&serverTimezone=UTC",
                ConfigManager.getMysqlHost(), ConfigManager.getMysqlPort(), ConfigManager.getMysqlDatabase());
        mysqlUsername = ConfigManager.getMysqlUsername();
        mysqlPassword = ConfigManager.getMysqlPassword();
        mysqlTable = ConfigManager.getMysqlTable();
    }
    
    private void initializeRedis() {
        try {
            JedisPoolConfig config = new JedisPoolConfig();
            config.setMaxTotal(10);
            config.setMaxIdle(5);
            config.setMinIdle(1);
            config.setTestOnBorrow(true);
            config.setTestOnReturn(true);
            
            if (redisPassword != null && !redisPassword.trim().isEmpty()) {
                jedisPool = new JedisPool(config, redisHost, redisPort, redisTimeout, redisPassword, redisDatabase);
            } else {
                jedisPool = new JedisPool(config, redisHost, redisPort, redisTimeout);
            }
            
            // 测试连接
            try (Jedis jedis = jedisPool.getResource()) {
                jedis.ping();
                logger.info("Redis connection established successfully");
            }
        } catch (Exception e) {
            logger.warn("Failed to initialize Redis connection: {}", e.getMessage());
            jedisPool = null;
        }
    }
    
    private void initializeMysql() {
        try {
            Class.forName("com.mysql.cj.jdbc.Driver");
            mysqlConnection = DriverManager.getConnection(mysqlUrl, mysqlUsername, mysqlPassword);
            
            // 创建表（如果不存在）
            createTableIfNotExists();
            
            logger.info("MySQL connection established successfully");
        } catch (Exception e) {
            logger.warn("Failed to initialize MySQL connection: {}", e.getMessage());
            mysqlConnection = null;
        }
    }
    
    private void createTableIfNotExists() {
        String createTableSql = String.format(
            "CREATE TABLE IF NOT EXISTS %s (" +
            "id BIGINT AUTO_INCREMENT PRIMARY KEY, " +
            "vin VARCHAR(17) NOT NULL, " +
            "kafka_timestamp BIGINT NOT NULL, " +
            "sig_timestamp BIGINT NOT NULL, " +
            "anomaly_type VARCHAR(50) NOT NULL, " +
            "severity VARCHAR(20) NOT NULL, " +
            "description TEXT, " +
            "sig_name VARCHAR(50) NOT NULL, " +
            "current_value TEXT, " +
            "threshold_value TEXT, " +
            "additional_info TEXT, " +
            "created_at DATETIME DEFAULT CURRENT_TIMESTAMP, " +
            "INDEX idx_vin (vin), " +
            "INDEX idx_kafka_timestamp (kafka_timestamp), " +
            "INDEX idx_sig_timestamp (sig_timestamp), " +
            "INDEX idx_anomaly_type (anomaly_type), " +
            "INDEX idx_severity (severity), " +
            "INDEX idx_sig_name (sig_name)" +
            ")", mysqlTable
        );
        
        try (Statement stmt = mysqlConnection.createStatement()) {
            stmt.execute(createTableSql);
            logger.info("Table {} created or already exists", mysqlTable);
        } catch (SQLException e) {
            logger.error("Failed to create table {}: {}", mysqlTable, e.getMessage());
        }
    }
    
    @Override
    public void invoke(AnomalyData anomalyData, Context context) throws Exception {
        if (anomalyData == null) {
            return;
        }
        
        logger.debug("Processing anomaly data: {}", anomalyData);
        
        // 优先尝试写入Redis
        boolean redisSuccess = writeToRedis(anomalyData);
        
        if (!redisSuccess) {
            // Redis写入失败，写入MySQL
            writeToMysql(anomalyData);
        }
    }
    
    private boolean writeToRedis(AnomalyData anomalyData) {
        if (jedisPool == null) {
            logger.warn("Redis connection not available");
            return false;
        }
        
        try (Jedis jedis = jedisPool.getResource()) {
            String jsonData = objectMapper.writeValueAsString(anomalyData);
            jedis.lpush(redisQueueKey, jsonData);
            logger.debug("Successfully wrote anomaly data to Redis queue: {}", anomalyData.getVin());
            return true;
        } catch (Exception e) {
            logger.error("Failed to write to Redis: {}", e.getMessage());
            return false;
        }
    }
    
    private void writeToMysql(AnomalyData anomalyData) {
        if (mysqlConnection == null) {
            logger.error("MySQL connection not available, data lost: {}", anomalyData.getVin());
            return;
        }
        
        String insertSql = String.format(
            "INSERT INTO %s (vin, kafka_timestamp, sig_timestamp, anomaly_type, severity, description, sig_name, current_value, threshold_value, additional_info) " +
            "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", mysqlTable
        );
        
        try (PreparedStatement stmt = mysqlConnection.prepareStatement(insertSql)) {
            stmt.setString(1, anomalyData.getVin());
            stmt.setLong(2, anomalyData.getKafkaTimestamp());
            stmt.setLong(3, anomalyData.getSigTimestamp());
            stmt.setString(4, anomalyData.getAnomalyType());
            stmt.setString(5, anomalyData.getSeverity());
            stmt.setString(6, anomalyData.getDescription());
            stmt.setString(7, anomalyData.getSigName());
            stmt.setString(8, anomalyData.getCurrentValue() != null ? anomalyData.getCurrentValue().toString() : null);
            stmt.setString(9, anomalyData.getThresholdValue() != null ? anomalyData.getThresholdValue().toString() : null);
            stmt.setString(10, anomalyData.getAdditionalInfo());
            
            int affected = stmt.executeUpdate();
            if (affected > 0) {
                logger.debug("Successfully wrote anomaly data to MySQL: {}", anomalyData.getVin());
            }
        } catch (SQLException e) {
            logger.error("Failed to write to MySQL: {}", e.getMessage());
        }
    }
    
    @Override
    public void close() throws Exception {
        super.close();
        
        if (jedisPool != null) {
            jedisPool.close();
            logger.info("Redis connection pool closed");
        }
        
        if (mysqlConnection != null && !mysqlConnection.isClosed()) {
            mysqlConnection.close();
            logger.info("MySQL connection closed");
        }
    }
}
