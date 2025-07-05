package com.vehicle.anomaly;

import com.vehicle.anomaly.config.ConfigManager;
import com.vehicle.anomaly.deserializer.VehicleDataDeserializer;
import com.vehicle.anomaly.model.AnomalyData;
import com.vehicle.anomaly.model.VehicleData;
import com.vehicle.anomaly.processor.AnomalyDetectionProcessor;
import com.vehicle.anomaly.sink.AnomalyDataSink;
import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.connector.kafka.source.KafkaSource;
import org.apache.flink.connector.kafka.source.enumerator.initializer.OffsetsInitializer;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.time.Duration;

/**
 * Flink Kafka处理器主类
 */
public class FlinkKafkaProcessor {
    
    private static final Logger logger = LoggerFactory.getLogger(FlinkKafkaProcessor.class);
    
    public static void main(String[] args) throws Exception {
        logger.info("Starting Flink Kafka Processor...");
        
        // 创建流处理环境
        final StreamExecutionEnvironment env = StreamExecutionEnvironment.getExecutionEnvironment();
        
        // 设置并行度
        env.setParallelism(ConfigManager.getFlinkParallelism());
        
        // 启用检查点
        env.enableCheckpointing(ConfigManager.getFlinkCheckpointInterval());
        
        // 创建Kafka源
        KafkaSource<VehicleData> kafkaSource = KafkaSource.<VehicleData>builder()
                .setBootstrapServers(ConfigManager.getKafkaBootstrapServers())
                .setTopics(ConfigManager.getKafkaTopic())
                .setGroupId(ConfigManager.getKafkaGroupId())
                .setStartingOffsets(OffsetsInitializer.latest())
                .setValueOnlyDeserializer(new VehicleDataDeserializer())
                .build();
        
        logger.info("Kafka source configured - Topic: {}, Group: {}, Servers: {}", 
                ConfigManager.getKafkaTopic(), 
                ConfigManager.getKafkaGroupId(), 
                ConfigManager.getKafkaBootstrapServers());
        
        // 从Kafka读取数据流
        DataStream<VehicleData> vehicleDataStream = env
                .fromSource(kafkaSource, WatermarkStrategy.<VehicleData>forBoundedOutOfOrderness(Duration.ofSeconds(20))
                        .withTimestampAssigner((event, timestamp) -> event.getSigTimestamp()), "Kafka Source")
                .name("Vehicle Data Stream");
        
        // 过滤空数据
        DataStream<VehicleData> filteredStream = vehicleDataStream
                .filter(data -> data != null && data.getSig() != null)
                .name("Filtered Vehicle Data");
        
        // 按VIN分组，然后进行异常检测处理
        DataStream<AnomalyData> anomalyDataStream = filteredStream
                .keyBy(VehicleData::getVin)
                .map(new AnomalyDetectionProcessor())
                .name("Anomaly Detection")
                .filter(anomaly -> anomaly != null)
                .name("Filtered Anomaly Data");
        
        // 输出到Redis/MySQL
        anomalyDataStream
                .addSink(new AnomalyDataSink())
                .name("Anomaly Data Sink");
        
        // 打印处理结果（用于调试）
        if (logger.isDebugEnabled()) {
            vehicleDataStream.print("Vehicle Data").setParallelism(1);
            anomalyDataStream.print("Anomaly Data").setParallelism(1);
        }
        
        logger.info("Flink job configured successfully");
        
        // 执行任务
        env.execute("Vehicle Anomaly Detection Job");
    }
}
