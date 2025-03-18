package router

import (
	"github.com/gin-gonic/gin"
)

// Router 接口定义了注册路由的方法
// 任何实现了 Route 方法的类型都可以被注册为一个路由
type Router interface {
	Route(r *gin.Engine)
}

// RegisterRouter 结构体用于注册路由
type RegisterRouter struct {
}

// New 创建并返回一个新的 RegisterRouter 实例
func New() *RegisterRouter {
	return &RegisterRouter{}
}

// Route 方法用于遍历所有已注册的路由并调用它们的 Route 方法
func (*RegisterRouter) Route(ro Router, r *gin.Engine) {
	ro.Route(r)
}

// routers 切片用于存储所有已注册的路由
var routers []Router

// InitRouter 方法用于初始化路由，将所有已注册的路由应用到 gin 的 Engine 对象上
func InitRouter(r *gin.Engine) {
	// 遍历所有已注册的路由并调用它们的 Route 方法
	for _, ro := range routers {
		ro.Route(r)
	}
}

// Register 方法用于向 routers 切片中注册一个或多个路由
func Register(ro ...Router) {
	routers = append(routers, ro...)
}
