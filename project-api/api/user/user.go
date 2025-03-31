package user

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"log"
	"net/http"
	"project-api/api/rpc"
	"project-api/pkg/model/user"
	common "project-common"
	"project-common/errs"
	"project-grpc/user/login"
	"time"
)

// HandlerUser 用户处理结构体
type HandlerUser struct {
}

// New 创建新的HandlerUser实例
func New() *HandlerUser {
	return &HandlerUser{}
}

// GetCaptcha 获取手机验证码
// 该方法负责处理获取验证码的请求，包括参数验证和验证码发送
func (*HandlerUser) GetCaptcha(ctx *gin.Context) {
	result := &common.Result{}
	//1. 获取参数
	mobile := ctx.PostForm("mobile")
	//2. 验证手机合法性
	if !common.VerifyMobile(mobile) {
		ctx.JSON(200, result.Fail(-1, "不合法"))
		return
	}
	//3. 生成验证码
	code := "123456"
	//4. 发送验证码（调用短信平台使用协程进行）
	go func() {
		time.Sleep(2 * time.Second)
		log.Println("调用短信平台发送短信")
		//发送成功 存入redis
		fmt.Println(mobile, code)
	}()
	ctx.JSON(200, result.Success("123456"))
}

// getCaptcha 为用户获取验证码。
// 该方法从请求上下文中提取手机号码，然后通过gRPC调用登录服务获取验证码。
// 它使用了context来设置操作的超时时间，以避免长时间运行的请求阻塞服务。
func (*HandlerUser) getCaptcha(ctx *gin.Context) {
	fmt.Println("hhhhhh")
	// 初始化结果对象，用于后续返回结果。
	result := &common.Result{}

	// 从请求中获取手机号码。
	mobile := ctx.PostForm("mobile")

	// 创建一个带有超时的context，以确保操作不会无限期地进行。
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 调用gRPC服务获取验证码。
	rsp, err := rpc.LoginServiceClient.GetCaptcha(c, &login.CaptchaMessage{Mobile: mobile})
	if err != nil {
		// 如果发生错误，解析gRPC错误以获取错误代码和消息。
		code, msg := errs.ParseGrpcError(err)
		// 返回错误响应。
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 如果成功，返回成功响应，包含验证码。
	ctx.JSON(http.StatusOK, result.Success(rsp.Code))
}

// register 用户注册函数
// 该函数处理用户的注册请求，包括参数验证、调用GRPC服务进行注册，并返回注册结果
func (u *HandlerUser) register(c *gin.Context) {
	// 初始化结果对象，用于后续返回API结果
	result := &common.Result{}

	// 定义注册请求参数模型，并尝试绑定请求参数到该模型
	var req user.RegisterReq
	err := c.ShouldBind(&req)
	if err != nil {
		// 如果参数绑定失败，返回错误信息
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数格式有误"))
		return
	}

	// 校验参数，确保参数值符合预期的业务规则
	if err := req.Verify(); err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	// 创建一个带有超时的context，用于调用GRPC服务
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 创建注册消息实例，并将请求参数复制到消息中，准备调用GRPC服务
	msg := &login.RegisterMessage{}
	// 使用Copier库将请求参数复制到消息中
	err = copier.Copy(msg, req)
	if err != nil {
		// 如果复制失败，返回错误信息
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "copy有误"))
		return
	}

	// 调用GRPC服务的Register方法，进行用户注册
	_, err = rpc.LoginServiceClient.Register(ctx, msg)
	if err != nil {
		// 如果注册失败，解析GRPC错误并返回
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 注册成功，返回成功信息
	c.JSON(http.StatusOK, result.Success(""))
}

// login 实现用户登录功能
// 该方法主要负责处理用户的登录请求，验证用户身份，并返回登录结果
func (u *HandlerUser) login(c *gin.Context) {
	//1.接收参数 参数模型
	// 初始化结果对象，用于后续返回API结果
	result := &common.Result{}
	// 定义登录请求模型，用于绑定用户登录的请求参数
	var req user.LoginReq
	// 将请求参数绑定到req对象中，如果参数格式有误，则返回错误信息
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数格式有误"))
		return
	}

	//2.调用user grpc 完成登录
	// 创建一个带有超时的上下文，以防止登录请求处理时间过长
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 初始化登录消息对象，用于调用gRPC服务
	msg := &login.LoginMessage{}
	// 将请求参数复制到登录消息对象中，如果复制有误，则返回错误信息
	err = copier.Copy(msg, req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "copy有误"))
		return
	}
	// TODO IP加入
	msg.Ip = GetIp(c)
	// 调用gRPC服务的Login方法进行登录，如果登录失败，则解析错误信息并返回
	loginRsp, err := rpc.LoginServiceClient.Login(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	// 初始化登录响应对象，用于处理登录结果
	rsp := &user.LoginRsp{}
	// 将登录结果复制到响应对象中，如果复制有误，则返回错误信息
	err = copier.Copy(rsp, loginRsp)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "copy有误"))
		return
	}

	//4.返回结果
	// 返回登录成功结果
	c.JSON(http.StatusOK, result.Success(rsp))
}

// myOrgList 处理用户获取自己所在的组织列表的请求。
// c *gin.Context: Gin框架的上下文对象，用于处理HTTP请求和响应。
func (u *HandlerUser) myOrgList(c *gin.Context) {
	// 初始化结果对象，用于构造响应结果。
	result := &common.Result{}

	// 从上下文中获取当前用户的memberId。
	memberIdStr, _ := c.Get("memberId")
	// 将memberId转换为int64类型。
	memberId := memberIdStr.(int64)

	req := &login.UserMessage{MemId: memberId}

	// 调用RPC服务，获取当前用户所在的组织列表。
	list, err2 := rpc.LoginServiceClient.MyOrgList(context.Background(), req)
	// 如果发生错误，解析gRPC错误并返回相应的错误响应。
	if err2 != nil {
		code, msg := errs.ParseGrpcError(err2)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 如果组织列表为空，返回空的组织列表。
	if list.OrganizationList == nil {
		c.JSON(http.StatusOK, result.Success([]*user.OrganizationList{}))
		return
	}

	// 初始化组织列表变量。
	var orgs []*user.OrganizationList
	// 将从RPC服务获取的组织列表复制到本地变量中。
	copier.Copy(&orgs, list.OrganizationList)
	// 返回成功的响应，包含组织列表。
	c.JSON(http.StatusOK, result.Success(orgs))
}

// GetIp 获取ip函数
func GetIp(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "::1" {
		ip = "127.0.0.1"
	}
	return ip
}
