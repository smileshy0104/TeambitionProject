package project

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"log"
	"project-api/config"
	"project-common/discovery"
	"project-common/logs"
	"project-grpc/account"
	"project-grpc/auth"
	"project-grpc/department"
	"project-grpc/project"
	"project-grpc/task"
)

// ProjectServiceClient 是一个全局变量，用于存储项目服务的 gRPC 客户端实例。
// 该客户端通过 InitRpcProjectClient 函数初始化，并在整个应用中使用。
var ProjectServiceClient project.ProjectServiceClient
var TaskServiceClient task.TaskServiceClient
var AccountServiceClient account.AccountServiceClient
var DepartmentServiceClient department.DepartmentServiceClient
var AuthServiceClient auth.AuthServiceClient

// InitRpcProjectClient 初始化项目服务的 gRPC 客户端。
// 该函数通过以下步骤完成初始化：
// 1. 创建一个基于 etcd 的自定义解析器，用于动态解析服务地址。
// 2. 注册自定义解析器到 gRPC 系统。
// 3. 使用解析器建立与项目服务的 gRPC 连接。
// 4. 如果连接成功，则创建并赋值项目服务的 gRPC 客户端实例。
func InitRpcProjectClient() {
	// 创建一个基于 etcd 的服务发现解析器，传入 etcd 地址和日志记录器。
	etcdRegister := discovery.NewResolver(config.C.EtcdConfig.Addrs, logs.LG)
	// 将自定义解析器注册到 gRPC 系统中。
	resolver.Register(etcdRegister)

	// 使用 etcd 解析器建立 gRPC 连接，目标服务名为 "project"。
	// 连接使用不安全的传输凭据（仅适用于开发环境，生产环境中应使用安全凭据）。
	conn, err := grpc.Dial("etcd:///project", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// 如果连接失败，记录致命错误并终止程序。
		log.Fatalf("无法连接到服务: %v", err)
	}
	// 使用已建立的 gRPC 连接创建项目服务的客户端实例。
	ProjectServiceClient = project.NewProjectServiceClient(conn)
	TaskServiceClient = task.NewTaskServiceClient(conn)
	AccountServiceClient = account.NewAccountServiceClient(conn)
	DepartmentServiceClient = department.NewDepartmentServiceClient(conn)
	AuthServiceClient = auth.NewAuthServiceClient(conn)
}
