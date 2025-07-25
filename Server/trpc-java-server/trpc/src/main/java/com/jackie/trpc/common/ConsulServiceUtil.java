package com.jackie.trpc.common;

import org.springframework.cloud.client.ServiceInstance;
import org.springframework.cloud.client.discovery.DiscoveryClient;
import org.springframework.stereotype.Component;
import io.netty.util.internal.ThreadLocalRandom;
import lombok.RequiredArgsConstructor;
import java.util.List;

//这个类从consul注册中心获取实例
@Component
@RequiredArgsConstructor
public class ConsulServiceUtil {
    private final DiscoveryClient discoveryClient;
    public String getServiceAddress(String serviceName) {
        // 获取所有实例（Consul客户端默认只返回健康实例）
        List<ServiceInstance> instances = discoveryClient.getInstances(serviceName);
        if (instances.isEmpty()) {
            throw new RuntimeException("Service not found: " + serviceName);
        }
        // 安全地随机选择一个实例（而不是固定选择第3个）
        int randomIndex = ThreadLocalRandom.current().nextInt(instances.size());
        ServiceInstance instance = instances.get(randomIndex);
        
        return instance.getHost() + ":" + instance.getPort();
    }
}