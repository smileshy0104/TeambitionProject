package midd

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"project-api/api/rpc"
	common "project-common"
	"project-common/errs"
	"project-grpc/user/login"
	"time"
)

// GetIp 获取ip函数
func GetIp(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "::1" {
		ip = "127.0.0.1"
	}
	return ip
}

// TokenVerify 返回一个中间件函数，用于验证请求中的Token
// 该中间件用于 Gin 框架的路由处理中，以实现Token验证的功能
func TokenVerifyOld() func(*gin.Context) {
	// 返回一个闭包函数，处理实际的Token验证逻辑
	return func(c *gin.Context) {
		// 初始化一个通用的结果对象，用于后续可能的响应
		result := &common.Result{}

		// 1. 从请求的header中获取Token
		token := c.GetHeader("Authorization")

		// 2. 调用user服务进行Token认证
		// 创建一个带有超时的context，以防止请求等待时间过长
		ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFunc() // 确保在函数退出时取消context

		// 调用RPC服务进行Token验证
		response, err := rpc.LoginServiceClient.TokenVerify(ctx, &login.LoginMessage{Token: token})
		if err != nil {
			// 如果发生错误，解析gRPC错误并返回相应的错误信息
			code, msg := errs.ParseGrpcError(err)
			c.JSON(http.StatusOK, result.Fail(code, msg))
			c.Abort() // 中止后续的路由处理
			return
		}

		// 3. 处理结果
		// 如果认证通过，将用户信息放入gin的上下文中，供后续处理函数使用
		// 如果认证失败，将不会执行后续的路由处理函数
		c.Set("memberId", response.Member.Id)
		c.Set("memberName", response.Member.Name)
		c.Set("organizationCode", response.Member.OrganizationCode)
		c.Next() // 继续执行后续的路由处理函数
	}
}

// TokenVerify 返回一个中间件函数，用于验证请求中的Token
// 该中间件用于 Gin 框架的路由处理中，以实现Token验证的功能
func TokenVerify() func(*gin.Context) {
	// 返回一个闭包函数，处理实际的Token验证逻辑
	return func(c *gin.Context) {
		// 初始化一个通用的结果对象，用于后续可能的响应
		result := &common.Result{}

		// 1. 从请求的header中获取Token
		token := c.GetHeader("Authorization")

		// 2. 调用user服务进行Token认证
		// 创建一个带有超时的context，以防止请求等待时间过长
		ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFunc() // 确保在函数退出时取消context
		// TODO 将Ip加入Token中，保证Token的安全性
		ip := GetIp(c)

		req := &login.LoginMessage{Token: token, Ip: ip}

		// 调用RPC服务进行Token验证
		response, err := rpc.LoginServiceClient.TokenVerify(ctx, req)
		if err != nil {
			// 如果发生错误，解析gRPC错误并返回相应的错误信息
			code, msg := errs.ParseGrpcError(err)
			c.JSON(http.StatusOK, result.Fail(code, msg))
			c.Abort() // 中止后续的路由处理
			return
		}

		// 3. 处理结果
		// 如果认证通过，将用户信息放入gin的上下文中，供后续处理函数使用
		// 如果认证失败，将不会执行后续的路由处理函数
		c.Set("memberId", response.Member.Id)
		c.Set("memberName", response.Member.Name)
		c.Set("organizationCode", response.Member.OrganizationCode)

		c.Next() // 继续执行后续的路由处理函数
	}
}
