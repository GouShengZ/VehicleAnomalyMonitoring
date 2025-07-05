package com.vehicle.anomaly.model;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * 车辆数据模型
 */
public class VehicleData {
    
    @JsonProperty("vin")
    private String vin;
    
    @JsonProperty("kafkaTimestamp")
    private long kafkaTimestamp;
    
    @JsonProperty("sigTimestamp")
    private long sigTimestamp;
    
    @JsonProperty("sig")
    private SignalData sig;
    
    // 构造函数
    public VehicleData() {}
    
    public VehicleData(String vin, long kafkaTimestamp, long sigTimestamp, SignalData sig) {
        this.vin = vin;
        this.kafkaTimestamp = kafkaTimestamp;
        this.sigTimestamp = sigTimestamp;
        this.sig = sig;
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
    
    public SignalData getSig() {
        return sig;
    }
    
    public void setSig(SignalData sig) {
        this.sig = sig;
    }
    
    @Override
    public String toString() {
        return "VehicleData{" +
                "vin='" + vin + '\'' +
                ", kafkaTimestamp=" + kafkaTimestamp +
                ", sigTimestamp=" + sigTimestamp +
                ", sig=" + sig +
                '}';
    }
}
