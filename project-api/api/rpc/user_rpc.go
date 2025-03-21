package rpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"log"
	"project-api/config"
	"project-common/discovery"
	"project-common/logs"
	"project-grpc/user/login"
)

var LoginServiceClient login.LoginServiceClient

func InitRpcUserClient() {
	etcdRegister := discovery.NewResolver(config.C.EtcdConfig.Addrs, logs.LG)
	resolver.Register(etcdRegister)

	conn, err := grpc.Dial("etcd:///user", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	LoginServiceClient = login.NewLoginServiceClient(conn)
}
