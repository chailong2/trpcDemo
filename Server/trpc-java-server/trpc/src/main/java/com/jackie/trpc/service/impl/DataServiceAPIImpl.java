package com.jackie.trpc.service.impl;

import com.jackie.trpc.service.DataServiceAPI;
import com.jackie.trpc.service.TrpcDemo;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.stereotype.Service;
import com.google.protobuf.InvalidProtocolBufferException;
import com.tencent.trpc.core.rpc.RpcContext;


@Service
public class DataServiceAPIImpl implements DataServiceAPI {

    @Autowired
    private RedisTemplate<String, byte[]> redisTemplate;

    @Override
    public TrpcDemo.mydata processData(RpcContext context, TrpcDemo.ProcessDataRequest request) {
        byte[] dataByte=redisTemplate.opsForValue().get(request.getRedisKey());
        TrpcDemo.mydata  mydata=null;
        try {
            mydata=TrpcDemo.mydata.parseFrom(dataByte);
            /**
             * 数据处理模拟操作
             */
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }
        return mydata;
    }
}
