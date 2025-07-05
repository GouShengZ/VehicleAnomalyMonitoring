package com.vehicle.anomaly.config;

import java.io.IOException;
import java.io.InputStream;
import java.util.Properties;

/**
 * 配置管理类
 */
public class ConfigManager {
    
    private static final String CONFIG_FILE = "application.properties";
    private static Properties properties;
    
    static {
        loadProperties();
    }
    
    private static void loadProperties() {
        properties = new Properties();
        try (InputStream input = ConfigManager.class.getClassLoader().getResourceAsStream(CONFIG_FILE)) {
            if (input != null) {
                properties.load(input);
            } else {
                throw new RuntimeException("Unable to find " + CONFIG_FILE);
            }
        } catch (IOException e) {
            throw new RuntimeException("Error loading configuration file", e);
        }
    }
    
    public static String get(String key) {
        return properties.getProperty(key);
    }
    
    public static String get(String key, String defaultValue) {
        return properties.getProperty(key, defaultValue);
    }
    
    public static int getInt(String key, int defaultValue) {
        String value = properties.getProperty(key);
        if (value != null) {
            try {
                return Integer.parseInt(value);
            } catch (NumberFormatException e) {
                return defaultValue;
            }
        }
        return defaultValue;
    }
    
    public static long getLong(String key, long defaultValue) {
        String value = properties.getProperty(key);
        if (value != null) {
            try {
                return Long.parseLong(value);
            } catch (NumberFormatException e) {
                return defaultValue;
            }
        }
        return defaultValue;
    }
    
    public static boolean getBoolean(String key, boolean defaultValue) {
        String value = properties.getProperty(key);
        if (value != null) {
            return Boolean.parseBoolean(value);
        }
        return defaultValue;
    }
    
    // Kafka配置
    public static String getKafkaBootstrapServers() {
        return get("kafka.bootstrap.servers", "localhost:9092");
    }
    
    public static String getKafkaGroupId() {
        return get("kafka.group.id", "vehicle-anomaly-group");
    }
    
    public static String getKafkaTopic() {
        return get("kafka.topic", "vehicle-data");
    }
    
    // Redis配置
    public static String getRedisHost() {
        return get("redis.host", "localhost");
    }
    
    public static int getRedisPort() {
        return getInt("redis.port", 6379);
    }
    
    public static String getRedisPassword() {
        return get("redis.password", "");
    }
    
    public static int getRedisDatabase() {
        return getInt("redis.database", 0);
    }
    
    public static int getRedisTimeout() {
        return getInt("redis.timeout", 5000);
    }
    
    public static String getRedisQueueKey() {
        return get("redis.queue.key", "vehicle_anomaly_queue");
    }
    
    // MySQL配置
    public static String getMysqlHost() {
        return get("mysql.host", "localhost");
    }
    
    public static int getMysqlPort() {
        return getInt("mysql.port", 3306);
    }
    
    public static String getMysqlDatabase() {
        return get("mysql.database", "vehicle_anomaly");
    }
    
    public static String getMysqlUsername() {
        return get("mysql.username", "root");
    }
    
    public static String getMysqlPassword() {
        return get("mysql.password", "password");
    }
    
    public static String getMysqlTable() {
        return get("mysql.table", "anomaly_data");
    }
    
    // Flink配置
    public static int getFlinkParallelism() {
        return getInt("flink.parallelism", 1);
    }
    
    public static long getFlinkCheckpointInterval() {
        return getLong("flink.checkpoint.interval", 60000);
    }
}
