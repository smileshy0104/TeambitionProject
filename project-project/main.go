package main

import (
	"github.com/gin-gonic/gin"
	srv "project-common"
	"project-project/config"
	"project-project/router"
)

func main() {
	r := gin.Default()
	//路由
	router.InitRouter(r)
	//grpc服务注册
	gc := router.RegisterGrpc()
	//grpc服务注册到etcd
	router.RegisterEtcdServer()
	//初始化kafka
	c := config.InitKafkaWriter()
	stop := func() {
		gc.Stop()
		c()
	}
	//初始化rpc调用
	router.InitUserRpc()
	srv.Run(r, config.C.SC.Name, config.C.SC.Addr, stop)
}
