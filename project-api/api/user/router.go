// Package user 用户模块的路由定义
package user

// 导入必要的库
import (
	"github.com/gin-gonic/gin"
	"log"
	"project-api/api/midd"
	"project-api/api/rpc"
	"project-api/router"
)

// RouterUser 结构体定义用户相关的路由方法
type RouterUser struct {
}

// init 函数在用户模块加载时初始化，注册用户路由
func init() {
	log.Println("init user router")
	ru := &RouterUser{}
	router.Register(ru)
}

// Route 方法为用户模块定义路由规则
func (*RouterUser) Route(r *gin.Engine) {
	//初始化grpc的客户端连接
	rpc.InitRpcUserClient()
	h := New()
	// 定义登录验证码获取的API路由，使用POST方法
	r.POST("/project/login/getCaptcha", h.getCaptcha)
	// 定义用户注册的API路由，使用POST方法
	r.POST("/project/login/register", h.register)
	// 定义用户登录的API路由，使用POST方法
	r.POST("/project/login", h.login)
	// 定义用户退出登录的API路由，使用POST方法
	org := r.Group("/project/organization")
	// 使用TokenVerify中间件对组织列表的API进行身份验证
	org.Use(midd.TokenVerify())
	org.POST("/_getOrgList", h.myOrgList)
}
