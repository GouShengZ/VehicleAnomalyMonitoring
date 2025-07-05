package com.vehicle.anomaly.model;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * 异常数据模型 - 用于存储检测到的车辆异常事件
 */
public class AnomalyData {
    
    @JsonProperty("vin")
    private String vin;
    
    @JsonProperty("kafkaTimestamp")
    private long kafkaTimestamp;
    
    @JsonProperty("sigTimestamp")
    private long sigTimestamp;
    
    @JsonProperty("anomalyType")
    private String anomalyType;
    
    @JsonProperty("severity")
    private String severity;
    
    @JsonProperty("description")
    private String description;
    
    @JsonProperty("sigName")
    private String sigName;
    
    @JsonProperty("currentValue")
    private Object currentValue;
    
    @JsonProperty("thresholdValue")
    private Object thresholdValue;
    
    @JsonProperty("additionalInfo")
    private String additionalInfo;
    
    // 构造函数
    public AnomalyData() {}
    
    public AnomalyData(String vin, long kafkaTimestamp, long sigTimestamp, 
                      String anomalyType, String severity, String description, 
                      String sigName, Object currentValue, Object thresholdValue, 
                      String additionalInfo) {
        this.vin = vin;
        this.kafkaTimestamp = kafkaTimestamp;
        this.sigTimestamp = sigTimestamp;
        this.anomalyType = anomalyType;
        this.severity = severity;
        this.description = description;
        this.sigName = sigName;
        this.currentValue = currentValue;
        this.thresholdValue = thresholdValue;
        this.additionalInfo = additionalInfo;
    }
    
    // Getters and Setters
    public String getVin() {
        return vin;
    }
    
    public void setVin(String vin) {
        this.vin = vin;
    }
    
    public long getKafkaTimestamp() {
        return kafkaTimestamp;
    }
    
    public void setKafkaTimestamp(long kafkaTimestamp) {
        this.kafkaTimestamp = kafkaTimestamp;
    }
    
    public long getSigTimestamp() {
        return sigTimestamp;
    }
    
    public void setSigTimestamp(long sigTimestamp) {
        this.sigTimestamp = sigTimestamp;
    }
    
    public String getAnomalyType() {
        return anomalyType;
    }
    
    public void setAnomalyType(String anomalyType) {
        this.anomalyType = anomalyType;
    }
    
    public String getSeverity() {
        return severity;
    }
    
    public void setSeverity(String severity) {
        this.severity = severity;
    }
    
    public String getDescription() {
        return description;
    }
    
    public void setDescription(String description) {
        this.description = description;
    }
    
    public String getSigName() {
        return sigName;
    }
    
    public void setSigName(String sigName) {
        this.sigName = sigName;
    }
    
    public Object getCurrentValue() {
        return currentValue;
    }
    
    public void setCurrentValue(Object currentValue) {
        this.currentValue = currentValue;
    }
    
    public Object getThresholdValue() {
        return thresholdValue;
    }
    
    public void setThresholdValue(Object thresholdValue) {
        this.thresholdValue = thresholdValue;
    }
    
    public String getAdditionalInfo() {
        return additionalInfo;
    }
    
    public void setAdditionalInfo(String additionalInfo) {
        this.additionalInfo = additionalInfo;
    }
    
    @Override
    public String toString() {
        return "AnomalyData{" +
                "vin='" + vin + '\'' +
                ", kafkaTimestamp=" + kafkaTimestamp +
                ", sigTimestamp=" + sigTimestamp +
                ", anomalyType='" + anomalyType + '\'' +
                ", severity='" + severity + '\'' +
                ", description='" + description + '\'' +
                ", sigName='" + sigName + '\'' +
                ", currentValue=" + currentValue +
                ", thresholdValue=" + thresholdValue +
                ", additionalInfo='" + additionalInfo + '\'' +
                '}';
    }
}
