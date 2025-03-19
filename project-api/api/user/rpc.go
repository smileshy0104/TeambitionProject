package user

import (
	loginServiceV1 "project-user/pkg/service/login.service.v1"
)

var LoginServiceClient loginServiceV1.LoginServiceClient

func InitRpcUserClient() {
	//etcdRegister := discovery.NewResolver(config.C.EtcdConfig.Addrs, logs.LG)
	//resolver.Register(etcdRegister)
	//
	//conn, err := grpc.Dial("etcd:///user", grpc.WithTransportCredentials(insecure.NewCredentials()))
	//if err != nil {
	//	log.Fatalf("did not connect: %v", err)
	//}
	//LoginServiceClient = loginServiceV1.NewLoginServiceClient(conn)
}
