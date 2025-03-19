package router

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"log"
	"net"
	"project-common/discovery"
	"project-common/logs"
	"project-user/config"
	loginServiceV1 "project-user/pkg/service/login.service.v1"
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

type gRPCConfig struct {
	Addr         string
	RegisterFunc func(*grpc.Server)
}

func RegisterGrpc() *grpc.Server {
	c := gRPCConfig{
		Addr: config.C.GC.Addr,
		RegisterFunc: func(g *grpc.Server) {
			loginServiceV1.RegisterLoginServiceServer(g, loginServiceV1.New())
		}}
	s := grpc.NewServer()
	c.RegisterFunc(s)
	lis, err := net.Listen("tcp", c.Addr)
	if err != nil {
		log.Println("cannot listen")
	}
	go func() {
		log.Printf("grpc server started as: %s \n", c.Addr)
		err = s.Serve(lis)
		if err != nil {
			log.Println("server started error", err)
			return
		}
	}()
	return s
}

func RegisterEtcdServer() {
	etcdRegister := discovery.NewResolver(config.C.EtcdConfig.Addrs, logs.LG)
	resolver.Register(etcdRegister)

	info := discovery.Server{
		Name:    config.C.GC.Name,
		Addr:    config.C.GC.Addr,
		Version: config.C.GC.Version,
		Weight:  config.C.GC.Weight,
	}
	r := discovery.NewRegister(config.C.EtcdConfig.Addrs, logs.LG)
	_, err := r.Register(info, 2)
	if err != nil {
		log.Fatalln(err)
	}
}
