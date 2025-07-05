package com.vehicle.anomaly.util;

import com.vehicle.anomaly.config.ConfigManager;
import com.vehicle.anomaly.model.SignalData;
import com.vehicle.anomaly.model.VehicleData;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.kafka.clients.producer.KafkaProducer;
import org.apache.kafka.clients.producer.ProducerConfig;
import org.apache.kafka.clients.producer.ProducerRecord;
import org.apache.kafka.common.serialization.StringSerializer;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Properties;
import java.util.Random;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

/**
 * 测试数据生成器 - 用于向Kafka发送测试数据
 */
public class TestDataGenerator {
    
    private static final Logger logger = LoggerFactory.getLogger(TestDataGenerator.class);
    
    private final KafkaProducer<String, String> producer;
    private final ObjectMapper objectMapper;
    private final Random random;
    private final ScheduledExecutorService scheduler;
    
    public TestDataGenerator() {
        this.objectMapper = new ObjectMapper();
        this.random = new Random();
        this.scheduler = Executors.newScheduledThreadPool(1);
        
        // 配置Kafka生产者
        Properties props = new Properties();
        props.put(ProducerConfig.BOOTSTRAP_SERVERS_CONFIG, ConfigManager.getKafkaBootstrapServers());
        props.put(ProducerConfig.KEY_SERIALIZER_CLASS_CONFIG, StringSerializer.class.getName());
        props.put(ProducerConfig.VALUE_SERIALIZER_CLASS_CONFIG, StringSerializer.class.getName());
        props.put(ProducerConfig.ACKS_CONFIG, "all");
        props.put(ProducerConfig.RETRIES_CONFIG, 3);
        props.put(ProducerConfig.BATCH_SIZE_CONFIG, 16384);
        props.put(ProducerConfig.LINGER_MS_CONFIG, 1);
        props.put(ProducerConfig.BUFFER_MEMORY_CONFIG, 33554432);
        
        this.producer = new KafkaProducer<>(props);
    }
    
    /**
     * 生成随机车辆数据
     */
    private VehicleData generateVehicleData() {
        String[] vins = {"1HGCM82633A123456", "1HGCM82633A123457", "1HGCM82633A123458", "1HGCM82633A123459", "1HGCM82633A123460"};
        String vin = vins[random.nextInt(vins.length)];
        
        long currentTime = System.currentTimeMillis();
        long kafkaTimestamp = currentTime;
        long sigTimestamp = currentTime - random.nextInt(1000); // 信号时间戳稍早于Kafka时间戳
        
        // 随机生成不同类型的信号
        SignalData signalData = generateRandomSignal();
        
        return new VehicleData(vin, kafkaTimestamp, sigTimestamp, signalData);
    }
    
    /**
     * 生成随机信号数据
     */
    private SignalData generateRandomSignal() {
        String[] signalTypes = {"speed", "AEB", "NOA_EXIT", "ACC_EXIT", "LKA_EXIT"};
        String sigName = signalTypes[random.nextInt(signalTypes.length)];
        
        Object value;
        
        switch (sigName) {
            case "speed":
                // 生成速度信号，有一定概率产生异常速度变化
                value = generateSpeedValue();
                break;
            case "AEB":
                // AEB信号，10%概率触发
                value = random.nextDouble() < 0.1;
                break;
            case "NOA_EXIT":
                // NOA退出信号，5%概率触发
                value = random.nextDouble() < 0.05;
                break;
            case "ACC_EXIT":
                // ACC退出信号，8%概率触发
                value = random.nextDouble() < 0.08;
                break;
            case "LKA_EXIT":
                // LKA退出信号，6%概率触发
                value = random.nextDouble() < 0.06;
                break;
            default:
                value = random.nextDouble() * 100;
        }
        
        return new SignalData(sigName, value);
    }
    
    /**
     * 生成速度值，有一定概率产生异常变化
     */
    private double generateSpeedValue() {
        // 正常速度范围 0-30 m/s (0-108 km/h)
        double normalSpeed = random.nextDouble() * 30;
        
        // 10%概率产生异常速度变化
        if (random.nextDouble() < 0.1) {
            // 产生异常速度变化（超过阈值）
            double anomalySpeed = normalSpeed + (random.nextBoolean() ? 1 : -1) * (20 + random.nextDouble() * 10);
            return Math.max(0, anomalySpeed);
        }
        
        return normalSpeed;
    }
    
    /**
     * 发送数据到Kafka
     */
    private void sendData() {
        try {
            VehicleData vehicleData = generateVehicleData();
            String jsonData = objectMapper.writeValueAsString(vehicleData);
            
            ProducerRecord<String, String> record = new ProducerRecord<>(
                ConfigManager.getKafkaTopic(),
                vehicleData.getVin(),
                jsonData
            );
            
            producer.send(record, (metadata, exception) -> {
                if (exception != null) {
                    logger.error("Failed to send data to Kafka: {}", exception.getMessage());
                } else {
                    logger.debug("Data sent successfully: partition={}, offset={}", 
                        metadata.partition(), metadata.offset());
                }
            });
            
            logger.info("Generated and sent vehicle data: VIN={}, Signal={}, Value={}", 
                vehicleData.getVin(), vehicleData.getSig().getSigName(), vehicleData.getSig().getValue());
        } catch (Exception e) {
            logger.error("Error generating test data: {}", e.getMessage());
        }
    }
    
    /**
     * 开始生成测试数据
     */
    public void startGenerating(int intervalSeconds) {
        logger.info("Starting test data generation with interval: {} seconds", intervalSeconds);
        
        scheduler.scheduleAtFixedRate(
            this::sendData,
            0,
            intervalSeconds,
            TimeUnit.SECONDS
        );
    }
    
    /**
     * 停止生成测试数据
     */
    public void stop() {
        logger.info("Stopping test data generation...");
        scheduler.shutdown();
        producer.close();
    }
    
    public static void main(String[] args) {
        TestDataGenerator generator = new TestDataGenerator();
        
        // 每3秒生成一条数据
        generator.startGenerating(3);
        
        // 运行时间：120秒
        try {
            Thread.sleep(120000);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
        
        generator.stop();
        logger.info("Test data generation completed");
    }
}
