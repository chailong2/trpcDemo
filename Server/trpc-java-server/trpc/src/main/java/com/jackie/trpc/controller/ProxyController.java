package com.jackie.trpc.controller;

import com.jackie.trpc.service.TrpcDemo;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import com.google.protobuf.util.JsonFormat;
import com.jackie.trpc.common.SnowflakeIdGenerator;
import com.jackie.trpc.config.TrpcClientFactory;
import com.jackie.trpc.service.Mydata;
import com.tencent.trpc.core.rpc.RpcClientContext;
import com.tencent.trpc.core.rpc.RpcContext;

@RestController
@RequestMapping("/")
public class ProxyController {

    @Autowired
    TrpcClientFactory factory;

    @Autowired
    private RedisTemplate<String, byte[]>  redisTemplate;

    @Autowired
    private SnowflakeIdGenerator idGenerator;

    private static final String MYDATA_KEY_PREFIX = "mydata:";

    @RequestMapping("/process")
    public String sayHello(@RequestBody String jsonData) {
        Mydata mydata =null;
        try {
            Mydata.Builder builder = Mydata.newBuilder();
            JsonFormat.parser().merge(jsonData, builder);
            mydata = builder.build();
            long snowflakeId = idGenerator.nextId();
            String key = MYDATA_KEY_PREFIX + snowflakeId;
            redisTemplate.opsForValue().set(key, mydata.toByteArray());
            RpcContext ctx = new RpcClientContext();
            TrpcDemo.ProcessDataRequest param = TrpcDemo.ProcessDataRequest.newBuilder().setRedisKey(key).build();
            factory.createProxy(com.jackie.trpc.service.DataServiceAPI.class, "trpc.test.helloworld.DataService").processData(ctx, param);
        } catch (Exception e) {
            throw new RuntimeException("Redis operation failed", e);
        }
        return mydata.toString();
    }
}
