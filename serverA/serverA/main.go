package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/sony/sonyflake"
	"google.golang.org/protobuf/proto"
	"net"
	"net/http"
	"os"
	"strconv"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/log"
	pb "trpcdemo/proto"

	redis "github.com/go-redis/redis/v8"
)

var (
	//redis的地址
	redisAddr = getenv("REDIS_ADDR", "127.0.0.1:6379")
	//consul地址
	consulAddr = getenv("CONSUL_ADDR", "127.0.0.1:8500")
	//http服务器监听地址和端口
	httPort = getenv("HTTP_PORT", "9001")
	//要暴露的服务ip地址
	hostIp = getenv("HOST_IP", "127.0.0.1") //docker宿主机ip
)

func getenv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
func init() {

}

var redisClient *redis.Client

func main() {
	// 创建redis链接
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	//初始化TRpc服务
	s := trpc.NewServer()
	//暴露自身服务
	pb.RegisterDataServiceService(s, &DataService{})
	cfg := trpc.GlobalConfig()
	portStr := cfg.Server.Service[0].Port
	port, err := strconv.Atoi(strconv.Itoa(int(portStr)))
	// 注册到Consul
	log.Info("正在注册服务到Consul....")
	err = RegisterToConsul("trpc.test.helloworld.DataService", port, hostIp)
	if err != nil {
		panic(fmt.Sprintf("注册到Consul失败: %v", err))
	}
	go func() {
		log.Info("监听RPC端口:", port)
		if err := s.Serve(); err != nil {
			log.Fatal(err)
		}
	}()
	// 实现http接口处理
	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		//从consul中获取服务
		equipmentClient, err := NewEquipmentDataClient(consulAddr) // Consul地址
		if err != nil {
			log.Fatalf("创建不了代理服务: %v", err)
			return
		}
		log.Info("正在从Consul获取服务....")
		address, err := equipmentClient.consulSD.GetServiceAddress(equipmentClient.serviceName)
		if err != nil || address == "" { // 增加 address == "" 的判断
			address, err = equipmentClient.consulSD.GetServiceAddress(equipmentClient.serviceName)
			// 循环体或其他逻辑
		}
		log.Info("服务ip地址:", address)
		// 创建TRPC客户端
		c := pb.NewDataServiceClientProxy(client.WithTarget("ip://" + address))
		log.Info("接收到请求: %s", r.URL.Path)
		defer r.Body.Close()
		// 判断是否是POST方法
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		log.Info("解析数据中......")
		// 读取Data(解析json格式)
		req := pb.Mydata{}
		jr := json.NewDecoder(r.Body)
		err = jr.Decode(&req)
		if err != nil {
			http.Error(w, "Failed to parse body", http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("null"))
			return
		}
		// 使用雪花算法生成key
		log.Info("RedisKey生成中......")
		flake := sonyflake.NewSonyflake(sonyflake.Settings{})
		id, err := flake.NextID()
		if err != nil {
			panic(err)
		}
		key := fmt.Sprintf("process_data_%d", id)
		// 序列化body
		value, _ := proto.Marshal(&req)
		// 写入Redis
		ctx := context.Background()
		err = redisClient.Set(ctx, key, value, 0).Err()
		if err != nil {
			http.Error(w, "Failed to save to redis", http.StatusInternalServerError)
			return
		}
		log.Info("Redis写入成功")
		// 调用B服务rpc接口传入redis-key
		resp, err := c.ProcessData(context.Background(), &pb.ProcessDataRequest{RedisKey: key})
		if err != nil {
			http.Error(w, "Failed to call rpc of B", http.StatusInternalServerError)
			return
		}

		log.Info("调用RPC B成功: %s", key)

		// 返回数据
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resp.String()))
	})

	// 监听http端口
	log.Info("监听http:", httPort)
	http.ListenAndServe(":"+httPort, nil)
}

type DataService struct{}

func (g DataService) ProcessData(ctx context.Context, req *pb.ProcessDataRequest) (*pb.Mydata, error) {
	log.Info("接收到RPC请求: %s", req.RedisKey)

	val, err := redisClient.Get(ctx, req.RedisKey).Result()
	if err != nil {
		log.Error("读取redis出错:%v", err)
		return nil, err
	}
	// 直接反序列化 Protobuf
	data := pb.Mydata{}
	if err := proto.Unmarshal([]byte(val), &data); err != nil {
		log.Error("解析 Protobuf 数据出错:%v", err)
		return nil, err
	}

	log.Info("解析redis数据成功: %s", req.RedisKey)
	return &data, nil
}

// consul相关的函数
type ConsulServiceDiscovery struct {
	client *api.Client
}

type EquipmentDataClient struct {
	serviceName string
	consulSD    *ConsulServiceDiscovery
}

func NewConsulServiceDiscovery(consulAddr string) (*ConsulServiceDiscovery, error) {
	config := api.DefaultConfig()
	config.Address = consulAddr
	c, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &ConsulServiceDiscovery{client: c}, nil
}
func NewEquipmentDataClient(consulAddr string) (*EquipmentDataClient, error) {
	sd, err := NewConsulServiceDiscovery(consulAddr)
	if err != nil {
		return nil, err
	}
	return &EquipmentDataClient{
		serviceName: "equipment-DataService",
		consulSD:    sd,
	}, nil
}

func (c *ConsulServiceDiscovery) GetServiceAddress(serviceName string) (string, error) {
	entries, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", err
	}

	if len(entries) == 0 {
		return "", fmt.Errorf("no healthy instances found for service: %s", serviceName)
	}

	// 简单的负载均衡：随机选择一个健康实例
	selected := entries[0]
	address := fmt.Sprintf("%s:%d", selected.Service.Address, selected.Service.Port)
	return address, nil
}

//注册服务到consul

func RegisterToConsul(serviceName string, port int, ip string) error {
	// 创建Consul客户端配置
	config := api.DefaultConfig()
	config.Address = consulAddr
	// 创建客户端
	client, err := api.NewClient(config)
	if err != nil {
		return fmt.Errorf("创建Consul客户端失败: %v", err)
	}
	// 创建服务注册信息
	registration := &api.AgentServiceRegistration{
		ID:      serviceName + "-" + ip + "-" + strconv.Itoa(port), // 服务唯一ID
		Name:    serviceName,                                       // 服务名称
		Address: ip,                                                // 服务IP
		Port:    port,                                              // 服务端口
		Check: &api.AgentServiceCheck{
			TCP:      net.JoinHostPort(ip, strconv.Itoa(port)), // 健康检查地址
			Interval: "10s",                                    // 健康检查间隔
			Timeout:  "5s",                                     // 超时时间
		},
	}
	// 注册服务
	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("注册服务到Consul失败: %v", err)
	}
	return nil
}
