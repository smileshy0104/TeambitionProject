package main

import (
	"github.com/gin-gonic/gin"
	_ "project-api/api"
	"project-api/config"
	"project-api/router"
	srv "project-common"
)

func main() {
	r := gin.Default()
	//路由
	router.InitRouter(r)
	srv.Run(r, config.C.SC.Name, config.C.SC.Addr, nil)
}
