package com.vehicle.anomaly.model;

import java.util.LinkedList;
import java.util.Queue;

/**
 * 速度状态模型 - 用于存储每个VIN的历史速度数据
 */
public class SpeedState {
    
    private final Queue<SpeedRecord> speedHistory;
    private final long timeWindowMs;
    
    public SpeedState(long timeWindowMs) {
        this.speedHistory = new LinkedList<>();
        this.timeWindowMs = timeWindowMs;
    }
    
    /**
     * 添加速度记录
     */
    public void addSpeedRecord(long timestamp, double speed) {
        speedHistory.offer(new SpeedRecord(timestamp, speed));
        // 清理过期数据
        cleanExpiredRecords(timestamp);
    }
    
    /**
     * 清理过期数据
     */
    private void cleanExpiredRecords(long currentTimestamp) {
        while (!speedHistory.isEmpty() && 
               currentTimestamp - speedHistory.peek().getTimestamp() > timeWindowMs) {
            speedHistory.poll();
        }
    }
    
    /**
     * 获取指定时间范围内的速度记录
     */
    public Queue<SpeedRecord> getSpeedRecordsInRange(long currentTimestamp, long rangeMs) {
        Queue<SpeedRecord> result = new LinkedList<>();
        long startTime = currentTimestamp - rangeMs;
        
        for (SpeedRecord record : speedHistory) {
            if (record.getTimestamp() >= startTime && record.getTimestamp() <= currentTimestamp) {
                result.offer(record);
            }
        }
        
        return result;
    }
    
    /**
     * 获取最新的速度记录
     */
    public SpeedRecord getLatestSpeedRecord() {
        if (speedHistory.isEmpty()) {
            return null;
        }
        
        SpeedRecord latest = null;
        for (SpeedRecord record : speedHistory) {
            if (latest == null || record.getTimestamp() > latest.getTimestamp()) {
                latest = record;
            }
        }
        return latest;
    }
    
    /**
     * 检查是否有足够的历史数据
     */
    public boolean hasHistoryData() {
        return !speedHistory.isEmpty();
    }
    
    /**
     * 获取历史记录数量
     */
    public int getHistorySize() {
        return speedHistory.size();
    }
    
    /**
     * 速度记录内部类
     */
    public static class SpeedRecord {
        private final long timestamp;
        private final double speed;
        
        public SpeedRecord(long timestamp, double speed) {
            this.timestamp = timestamp;
            this.speed = speed;
        }
        
        public long getTimestamp() {
            return timestamp;
        }
        
        public double getSpeed() {
            return speed;
        }
        
        @Override
        public String toString() {
            return "SpeedRecord{" +
                    "timestamp=" + timestamp +
                    ", speed=" + speed +
                    '}';
        }
    }
}
