package com.vehicle.anomaly.deserializer;

import com.vehicle.anomaly.model.VehicleData;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import org.apache.flink.api.common.serialization.DeserializationSchema;
import org.apache.flink.api.common.typeinfo.TypeInformation;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

/**
 * Kafka消息反序列化器
 */
public class VehicleDataDeserializer implements DeserializationSchema<VehicleData> {
    
    private static final Logger logger = LoggerFactory.getLogger(VehicleDataDeserializer.class);
    
    private transient ObjectMapper objectMapper;
    
    @Override
    public void open(InitializationContext context) throws Exception {
        objectMapper = new ObjectMapper();
        objectMapper.registerModule(new JavaTimeModule());
    }
    
    @Override
    public VehicleData deserialize(byte[] message) throws IOException {
        if (message == null || message.length == 0) {
            return null;
        }
        
        try {
            String jsonString = new String(message, StandardCharsets.UTF_8);
            logger.debug("Deserializing message: {}", jsonString);
            
            VehicleData vehicleData = objectMapper.readValue(jsonString, VehicleData.class);
            logger.debug("Successfully deserialized VehicleData: {}", vehicleData);
            
            return vehicleData;
        } catch (Exception e) {
            logger.error("Failed to deserialize message: {}", e.getMessage());
            // 返回null而不是抛出异常，让Flink继续处理其他消息
            return null;
        }
    }
    
    @Override
    public boolean isEndOfStream(VehicleData nextElement) {
        return false;
    }
    
    @Override
    public TypeInformation<VehicleData> getProducedType() {
        return TypeInformation.of(VehicleData.class);
    }
}
