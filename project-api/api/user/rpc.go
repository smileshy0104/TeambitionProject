package user

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"log"
	"project-api/config"
	"project-common/discovery"
	"project-common/logs"
	loginServiceV1 "project-grpc/user/login"
)

// LoginServiceClient 是一个全局变量，用于存储初始化后的 gRPC 客户端，
// 该客户端用于与用户登录服务进行通信。
var LoginServiceClient loginServiceV1.LoginServiceClient

// InitRpcUserClient 初始化 gRPC 用户登录服务客户端。
// 此函数通过以下步骤完成初始化：
// 1. 创建一个基于 ETCD 的自定义解析器，用于动态发现服务地址。
// 2. 使用 insecure 模式建立 gRPC 连接，并指定目标服务为 "etcd:///user"。
// 3. 如果连接成功，则创建并设置全局的 LoginServiceClient。
func InitRpcUserClient() {
	// 创建一个基于 ETCD 的自定义解析器，用于动态解析服务地址。
	// 参数包括 ETCD 地址列表和日志记录器。
	etcdRegister := discovery.NewResolver(config.C.EtcdConfig.Addrs, logs.LG)
	resolver.Register(etcdRegister)

	// 建立到目标服务的 gRPC 连接，使用 insecure 模式（不验证 TLS）。
	conn, err := grpc.Dial("etcd:///user", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// 如果连接失败，记录致命错误并终止程序运行。
		log.Fatalf("did not connect: %v", err)
	}

	// 使用已建立的连接创建 LoginServiceClient，并将其赋值给全局变量。
	LoginServiceClient = loginServiceV1.NewLoginServiceClient(conn)
}
