// Package user 用户模块的路由定义
package user

// 导入必要的库
import (
	"github.com/gin-gonic/gin"
	"log"
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
	h := New()
	// 定义登录验证码获取的API路由，使用POST方法
	r.POST("/project/login/getCaptcha", h.getCaptcha)
}
