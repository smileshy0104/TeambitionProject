package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	srv "project-common"
	"project-user/config"
	"project-user/router"
)

func main() {
	r := gin.Default()
	//路由
	router.InitRouter(r)
	//grpc服务注册
	gc := router.RegisterGrpc()
	fmt.Println("grpc服务注册成功")
	//grpc服务注册到etcd
	router.RegisterEtcdServer()
	stop := func() {
		gc.Stop()
	}
	srv.Run(r, config.C.SC.Name, config.C.SC.Addr, stop)
}
