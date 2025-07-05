package com.vehicle.anomaly.processor;

import com.vehicle.anomaly.model.AnomalyData;
import com.vehicle.anomaly.model.SpeedState;
import com.vehicle.anomaly.model.VehicleData;
import org.apache.flink.api.common.functions.RichMapFunction;
import org.apache.flink.api.common.state.ValueState;
import org.apache.flink.api.common.state.ValueStateDescriptor;
import org.apache.flink.api.common.typeinfo.TypeHint;
import org.apache.flink.api.common.typeinfo.TypeInformation;
import org.apache.flink.configuration.Configuration;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Queue;

/**
 * 车辆数据异常检测处理器
 * 功能：
 * 1. 按照信号触发时间戳排序处理数据
 * 2. 速度异常检测（基于历史状态）
 * 3. 系统信号异常检测（AEB、NOA退出、ACC退出、LKA退出）
 */
public class AnomalyDetectionProcessor extends RichMapFunction<VehicleData, AnomalyData> {
    
    private static final Logger logger = LoggerFactory.getLogger(AnomalyDetectionProcessor.class);
    
    // 异常检测阈值
    private static final double SPEED_DIFF_1S = 18.0; // 1秒内速度差异阈值 (m/s)
    private static final double SPEED_DIFF_2S = 24.0; // 2秒内速度差异阈值 (m/s)
    private static final long TIME_WINDOW_MS = 30000; // 30秒时间窗口
    
    // 状态变量用于存储每个VIN的历史速度数据
    private transient ValueState<SpeedState> speedStateDescriptor;
    
    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        
        // 初始化状态描述符
        ValueStateDescriptor<SpeedState> descriptor = new ValueStateDescriptor<>(
            "speedState",
            TypeInformation.of(new TypeHint<SpeedState>() {})
        );
        
        speedStateDescriptor = getRuntimeContext().getState(descriptor);
        
        logger.info("AnomalyDetectionProcessor initialized");
    }
    
    @Override
    public AnomalyData map(VehicleData vehicleData) throws Exception {
        if (vehicleData == null || vehicleData.getSig() == null) {
            return null;
        }
        
        logger.debug("Processing vehicle data: {}", vehicleData);
        
        String sigName = vehicleData.getSig().getSigName();
        
        // 处理速度信号
        if ("speed".equals(sigName)) {
            return processSpeedSignal(vehicleData);
        }
        
        // 处理系统信号异常
        return processSystemSignal(vehicleData);
    }
    
    /**
     * 处理速度信号异常检测
     */
    private AnomalyData processSpeedSignal(VehicleData vehicleData) throws Exception {
        Double currentSpeed = vehicleData.getSig().getValueAs(Double.class);
        if (currentSpeed == null) {
            return null;
        }
        
        long currentTimestamp = vehicleData.getSigTimestamp();
        
        // 获取当前VIN的速度状态
        SpeedState speedState = speedStateDescriptor.value();
        if (speedState == null) {
            speedState = new SpeedState(TIME_WINDOW_MS);
        }
        
        // 检查速度异常
        AnomalyData anomaly = checkSpeedAnomaly(vehicleData, speedState, currentSpeed, currentTimestamp);
        
        // 更新速度状态
        speedState.addSpeedRecord(currentTimestamp, currentSpeed);
        speedStateDescriptor.update(speedState);
        
        return anomaly;
    }
    
    /**
     * 检查速度异常
     */
    private AnomalyData checkSpeedAnomaly(VehicleData vehicleData, SpeedState speedState, 
                                         double currentSpeed, long currentTimestamp) {
        
        // 获取1秒内的速度记录
        Queue<SpeedState.SpeedRecord> records1s = speedState.getSpeedRecordsInRange(currentTimestamp, 1000);
        
        // 检查1秒内速度变化
        for (SpeedState.SpeedRecord record : records1s) {
            double speedDiff = Math.abs(currentSpeed - record.getSpeed());
            if (speedDiff > SPEED_DIFF_1S) {
                return createAnomalyData(vehicleData, "SPEED_CHANGE_1S", "HIGH",
                    "Speed changed too rapidly within 1 second", 
                    speedDiff, SPEED_DIFF_1S,
                    String.format("Previous speed: %.2f m/s, Current speed: %.2f m/s, Time diff: %d ms",
                        record.getSpeed(), currentSpeed, currentTimestamp - record.getTimestamp()));
            }
        }
        
        // 获取2秒内的速度记录
        Queue<SpeedState.SpeedRecord> records2s = speedState.getSpeedRecordsInRange(currentTimestamp, 2000);
        
        // 检查2秒内速度变化
        for (SpeedState.SpeedRecord record : records2s) {
            double speedDiff = Math.abs(currentSpeed - record.getSpeed());
            if (speedDiff > SPEED_DIFF_2S) {
                return createAnomalyData(vehicleData, "SPEED_CHANGE_2S", "MEDIUM",
                    "Speed changed too rapidly within 2 seconds", 
                    speedDiff, SPEED_DIFF_2S,
                    String.format("Previous speed: %.2f m/s, Current speed: %.2f m/s, Time diff: %d ms",
                        record.getSpeed(), currentSpeed, currentTimestamp - record.getTimestamp()));
            }
        }
        
        return null;
    }
    
    /**
     * 处理系统信号异常检测
     */
    private AnomalyData processSystemSignal(VehicleData vehicleData) {
        String sigName = vehicleData.getSig().getSigName();
        Object value = vehicleData.getSig().getValue();
        
        // AEB信号触发
        if ("AEB".equals(sigName)) {
            Boolean aebActive = vehicleData.getSig().getValueAs(Boolean.class);
            if (aebActive != null && aebActive) {
                return createAnomalyData(vehicleData, "AEB_TRIGGERED", "HIGH",
                    "Autonomous Emergency Braking activated", 
                    value, true,
                    "AEB system intervention detected");
            }
        }
        
        // NOA退出信号
        if ("NOA_EXIT".equals(sigName)) {
            Boolean noaExit = vehicleData.getSig().getValueAs(Boolean.class);
            if (noaExit != null && noaExit) {
                return createAnomalyData(vehicleData, "NOA_EXIT", "MEDIUM",
                    "Navigate on Autopilot exited", 
                    value, true,
                    "NOA system disengaged");
            }
        }
        
        // ACC退出信号
        if ("ACC_EXIT".equals(sigName)) {
            Boolean accExit = vehicleData.getSig().getValueAs(Boolean.class);
            if (accExit != null && accExit) {
                return createAnomalyData(vehicleData, "ACC_EXIT", "MEDIUM",
                    "Adaptive Cruise Control exited", 
                    value, true,
                    "ACC system disengaged");
            }
        }
        
        // LKA退出信号
        if ("LKA_EXIT".equals(sigName)) {
            Boolean lkaExit = vehicleData.getSig().getValueAs(Boolean.class);
            if (lkaExit != null && lkaExit) {
                return createAnomalyData(vehicleData, "LKA_EXIT", "MEDIUM",
                    "Lane Keeping Assist exited", 
                    value, true,
                    "LKA system disengaged");
            }
        }
        
        return null;
    }
    
    /**
     * 创建异常数据对象
     */
    private AnomalyData createAnomalyData(VehicleData vehicleData, String anomalyType, 
                                         String severity, String description, 
                                         Object currentValue, Object thresholdValue,
                                         String additionalInfo) {
        return new AnomalyData(
            vehicleData.getVin(),
            vehicleData.getKafkaTimestamp(),
            vehicleData.getSigTimestamp(),
            anomalyType,
            severity,
            description,
            vehicleData.getSig().getSigName(),
            currentValue,
            thresholdValue,
            additionalInfo
        );
    }
}
