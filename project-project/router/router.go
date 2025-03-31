package router

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"log"
	"net"
	"project-common/discovery"
	"project-common/logs"
	"project-grpc/account"
	"project-grpc/auth"
	"project-grpc/department"
	"project-grpc/menu"
	"project-grpc/project"
	"project-grpc/task"
	"project-project/config"
	"project-project/internal/interceptor"
	"project-project/internal/rpc"
	account_service_v1 "project-project/pkg/service/account.service.v1"
	auth_service_v1 "project-project/pkg/service/auth.service.v1"
	department_service_v1 "project-project/pkg/service/department.service.v1"
	menu_service_v1 "project-project/pkg/service/menu.service.v1"
	project_service_v1 "project-project/pkg/service/project.service.v1"
	task_service_v1 "project-project/pkg/service/task.service.v1"
)

// Router 接口
type Router interface {
	Route(r *gin.Engine)
}

type RegisterRouter struct {
}

func New() *RegisterRouter {
	return &RegisterRouter{}
}

func (*RegisterRouter) Route(ro Router, r *gin.Engine) {
	ro.Route(r)
}

var routers []Router

func InitRouter(r *gin.Engine) {
	//rg := New()
	//rg.Route(&user.RouterUser{}, r)
	for _, ro := range routers {
		ro.Route(r)
	}
}

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
			project.RegisterProjectServiceServer(g, project_service_v1.New())
			task.RegisterTaskServiceServer(g, task_service_v1.New())
			account.RegisterAccountServiceServer(g, account_service_v1.New())
			department.RegisterDepartmentServiceServer(g, department_service_v1.New())
			auth.RegisterAuthServiceServer(g, auth_service_v1.New())
			menu.RegisterMenuServiceServer(g, menu_service_v1.New())
		}}
	// 创建grpc服务——注册缓存
	s := grpc.NewServer(interceptor.New().Cache())
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

func InitUserRpc() {
	rpc.InitRpcUserClient()
}
