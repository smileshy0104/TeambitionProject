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
	group := r.Group("/project")
	// 使用TokenVerify中间件对项目列表的API进行身份验证
	group.Use(midd.TokenVerify())
	group.POST("/index", h.index)                            // Index 获取项目的菜单列表
	group.POST("/project/selfList", h.myProjectList)         // myProjectList 获取用户自身项目列表请求
	group.POST("/project", h.myProjectList)                  // myProjectList 获取项目列表请求
	group.POST("/project_template", h.projectTemplate)       // projectTemplate 获取项目模板列表请求
	group.POST("/project/save", h.projectSave)               // projectSave 保存项目请求
	group.POST("/project/read", h.readProject)               // readProject 读取项目请求
	group.POST("/project/recycle", h.recycleProject)         // recycleProject 删除项目请求
	group.POST("/project/recovery", h.recoveryProject)       // recoveryProject 恢复项目请求
	group.POST("/project_collect/collect", h.collectProject) // collectProject 收藏项目请求
	group.POST("/project/edit", h.editProject)               // editProject 编辑项目请求
}
