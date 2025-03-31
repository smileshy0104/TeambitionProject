package interceptor

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"project-common/encrypts"
	"project-grpc/project"
	"project-project/internal/dao"
	"project-project/internal/repo"
	"time"
)

// CacheInterceptor 是一个缓存拦截器，用于在gRPC服务器中实现缓存功能。
// 它通过缓存请求的响应结果来减少对后端服务的调用次数，从而提高性能。
type CacheInterceptor struct {
	cache    repo.Cache     // 用于与缓存存储进行交互的接口
	cacheMap map[string]any // 映射gRPC方法到其对应的响应类型
}

// CacheRespOption 是缓存响应选项的结构体，用于配置缓存的行为。
type CacheRespOption struct {
	path   string        // gRPC方法的路径
	typ    any           // 响应值的类型
	expire time.Duration // 缓存的过期时间
}

// New 创建并返回一个新的CacheInterceptor实例。
// 它初始化了cacheMap，可以根据需要添加更多的方法和对应的响应类型。
func New() *CacheInterceptor {
	cacheMap := make(map[string]any)
	// 可以在此处添加更多的方法和对应的响应类型
	// 例如：cacheMap["/project.service.v1.ProjectService/FindProjectByMemId"] = &project.MyProjectResponse{}
	cacheMap["/project.service.v1.ProjectService/FindProjectByMemId"] = &project.MyProjectResponse{}
	return &CacheInterceptor{cache: dao.Rc, cacheMap: cacheMap}
}

// Cache 返回一个gRPC服务器选项，用于注册缓存拦截器。
// 该拦截器会在请求处理前检查缓存，如果存在则直接返回缓存结果，否则执行请求并缓存结果。
func (c *CacheInterceptor) Cache() grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 检查是否有对应的响应类型，如果没有则直接调用处理器
		respType := c.cacheMap[info.FullMethod]
		if respType == nil {
			return handler(ctx, req)
		}

		// 创建一个带有超时的上下文，防止缓存操作阻塞过久
		con, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// 生成缓存键
		marshal, _ := json.Marshal(req)
		cacheKey := encrypts.Md5(string(marshal))

		// 尝试从缓存中获取结果
		respJson, _ := c.cache.Get(con, info.FullMethod+"::"+cacheKey)
		if respJson != "" {
			// 如果缓存存在，则解析并返回缓存结果
			json.Unmarshal([]byte(respJson), &respType)
			zap.L().Info(info.FullMethod + " 使用了缓存")
			return respType, nil
		}

		// 如果缓存不存在，则调用处理器并缓存结果
		resp, err = handler(ctx, req)
		bytes, _ := json.Marshal(resp)
		c.cache.Put(con, info.FullMethod+"::"+cacheKey, string(bytes), 5*time.Minute)
		zap.L().Info(info.FullMethod + " 结果已缓存")
		return
	})
}
