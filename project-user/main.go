package main

import (
	"github.com/gin-gonic/gin"
	srv "project-common"
	"project-user/router"
)

func main() {
	r := gin.Default()
	//路由
	router.InitRouter(r)
	srv.Run(r, "project-user", ":8080", nil)
}
