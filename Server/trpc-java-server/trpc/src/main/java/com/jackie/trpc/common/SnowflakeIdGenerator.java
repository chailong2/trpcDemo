package com.jackie.trpc.common;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

//雪花算法
@Component
public class SnowflakeIdGenerator {
    private final long datacenterId; // 数据中心ID
    private final long machineId;    // 机器ID
    private long sequence = 0L;
    private long lastTimestamp = -1L;

    public SnowflakeIdGenerator(
            @Value("${snowflake.datacenter-id:1}") long datacenterId,
            @Value("${snowflake.machine-id:1}") long machineId) {
        this.datacenterId = datacenterId;
        this.machineId = machineId;
    }

    public synchronized long nextId() {
        long timestamp = System.currentTimeMillis();

        if (timestamp < lastTimestamp) {
            throw new RuntimeException("Clock moved backwards");
        }

        if (timestamp == lastTimestamp) {
            sequence = (sequence + 1) & 0xFFF;
            if (sequence == 0) {
                timestamp = tilNextMillis(lastTimestamp);
            }
        } else {
            sequence = 0L;
        }

        lastTimestamp = timestamp;

        return ((timestamp - 1288834974657L) << 22) 
                | (datacenterId << 17) 
                | (machineId << 12) 
                | sequence;
    }

    private long tilNextMillis(long lastTimestamp) {
        long timestamp = System.currentTimeMillis();
        while (timestamp <= lastTimestamp) {
            timestamp = System.currentTimeMillis();
        }
        return timestamp;
    }
}