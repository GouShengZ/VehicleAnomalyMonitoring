package com.vehicle.anomaly.model;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * 信号数据模型
 */
public class SignalData {
    
    @JsonProperty("sigName")
    private String sigName;
    
    @JsonProperty("value")
    private Object value;
    
    // 构造函数
    public SignalData() {}
    
    public SignalData(String sigName, Object value) {
        this.sigName = sigName;
        this.value = value;
    }
    
    // Getters and Setters
    public String getSigName() {
        return sigName;
    }
    
    public void setSigName(String sigName) {
        this.sigName = sigName;
    }
    
    public Object getValue() {
        return value;
    }
    
    public void setValue(Object value) {
        this.value = value;
    }
    
    /**
     * 获取值并转换为指定类型
     */
    public <T> T getValueAs(Class<T> clazz) {
        if (value == null) {
            return null;
        }
        
        if (clazz.isInstance(value)) {
            return clazz.cast(value);
        }
        
        // 处理数值类型转换
        if (clazz == Double.class || clazz == double.class) {
            if (value instanceof Number) {
                return clazz.cast(((Number) value).doubleValue());
            }
            if (value instanceof String) {
                try {
                    return clazz.cast(Double.parseDouble((String) value));
                } catch (NumberFormatException e) {
                    return null;
                }
            }
        }
        
        if (clazz == Integer.class || clazz == int.class) {
            if (value instanceof Number) {
                return clazz.cast(((Number) value).intValue());
            }
            if (value instanceof String) {
                try {
                    return clazz.cast(Integer.parseInt((String) value));
                } catch (NumberFormatException e) {
                    return null;
                }
            }
        }
        
        if (clazz == Boolean.class || clazz == boolean.class) {
            if (value instanceof Boolean) {
                return clazz.cast(value);
            }
            if (value instanceof String) {
                return clazz.cast(Boolean.parseBoolean((String) value));
            }
            if (value instanceof Number) {
                return clazz.cast(((Number) value).intValue() != 0);
            }
        }
        
        if (clazz == String.class) {
            return clazz.cast(value.toString());
        }
        
        return null;
    }
    
    @Override
    public String toString() {
        return "SignalData{" +
                "sigName='" + sigName + '\'' +
                ", value=" + value +
                '}';
    }
}
