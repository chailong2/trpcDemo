package com.jackie.trpc;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.client.discovery.EnableDiscoveryClient;
import com.tencent.trpc.spring.boot.starters.annotation.EnableTRpc;


@EnableTRpc
@SpringBootApplication
@EnableDiscoveryClient
public class TrpcApplication {
    public static void main(String[] args) {
        //ConfigManager.getInstance().start();
        SpringApplication.run(TrpcApplication.class, args);
    }
}
