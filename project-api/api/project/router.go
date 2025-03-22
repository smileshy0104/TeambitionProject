package project

import (
	"github.com/gin-gonic/gin"
	"log"
	"project-api/api/midd"
	"project-api/router"
)

type RouterProject struct {
}

func init() {
	log.Println("init project router")
	ru := &RouterProject{}
	router.Register(ru)
}

func (*RouterProject) Route(r *gin.Engine) {
	//初始化grpc的客户端连接
	//InitRpcProjectClient()
	h := New()
	// 定义对应的路由组规则
	group := r.Group("/project/index")
	// 使用TokenVerify中间件对项目列表的API进行身份验证
	group.Use(midd.TokenVerify())
	// 定义项目列表的API路由，使用POST方法
	group.POST("", h.index)
	// 定义对应的路由组规则
	group1 := r.Group("/project/project")
	// 使用TokenVerify中间件对项目列表的API进行身份验证
	group1.Use(midd.TokenVerify())
	group1.POST("/selfList", h.myProjectList)
}
