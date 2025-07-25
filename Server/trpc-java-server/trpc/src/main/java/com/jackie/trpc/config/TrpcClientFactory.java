package com.jackie.trpc.config;

import org.springframework.stereotype.Component;
import com.jackie.trpc.common.ConsulServiceUtil;
import com.tencent.trpc.core.common.config.BackendConfig;
import com.tencent.trpc.core.common.config.ConsumerConfig;

@Component
public class TrpcClientFactory {
    
    private final ConsulServiceUtil consulServiceUtil;
    
    public TrpcClientFactory(ConsulServiceUtil consulServiceUtil) {
        this.consulServiceUtil = consulServiceUtil;
    }
    
    public <T> T createProxy(Class<T> serviceInterface, String serviceName) {
        // 1. 创建Consumer配置
        ConsumerConfig<T> consumerConfig = new ConsumerConfig<>();
        consumerConfig.setServiceInterface(serviceInterface);
        // 2. 创建Backend配置（从Consul获取地址）
        BackendConfig backendConfig = new BackendConfig();
        String serviceAddress = consulServiceUtil.getServiceAddress(serviceName);
        backendConfig.setNamingUrl("ip://" + serviceAddress);
        backendConfig.setProtocol("trpc");
        backendConfig.setSerialization("pb");
        backendConfig.setNetwork("tcp");
        // 3. 创建并返回代理
        return backendConfig.getProxy(consumerConfig);
    }
}
