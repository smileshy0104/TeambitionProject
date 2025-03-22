package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/pkg/model"
	"project-api/pkg/model/pro"
	common "project-common"
	"project-common/errs"
	"project-grpc/project"
	"time"
)

type HandlerProject struct {
}

func New() *HandlerProject {
	return &HandlerProject{}
}

// index 处理项目索引请求。
// 该函数接收一个 gin.Context 参数，用于处理 HTTP 请求和响应。
// 它通过 ProjectServiceClient 调用 Index 方法获取索引数据，并返回给客户端。
func (p *HandlerProject) index(c *gin.Context) {
	// 初始化结果对象，用于后续构造响应。
	result := &common.Result{}

	// 创建一个带有超时的上下文，以确保请求不会无限期地等待。
	// 这里设置超时时间为2秒。
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	// 延迟取消上下文，以确保在函数退出时清理资源。
	defer cancel()

	// 创建一个空的 IndexMessage 对象，用于发送索引请求。
	msg := &project.IndexMessage{}

	// 调用 ProjectServiceClient 的 Index 方法获取索引数据。
	// 如果发生错误，解析 gRPC 错误并返回错误响应。
	indexResponse, err := ProjectServiceClient.Index(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 如果请求成功，返回成功响应，包含索引数据中的菜单信息。
	c.JSON(http.StatusOK, result.Success(indexResponse.Menus))
}

// myProjectList 处理我的项目列表请求
// 该方法通过gRPC调用获取项目列表，并返回给客户端
func (p *HandlerProject) myProjectList(c *gin.Context) {
	// 初始化结果对象
	result := &common.Result{}

	// 1. 获取参数
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 从上下文中获取memberId
	memberIdStr, _ := c.Get("memberId")
	memberId := memberIdStr.(int64)

	// 绑定分页参数
	page := &model.Page{}
	page.Bind(c)

	// 构建项目查询消息
	msg := &project.ProjectRpcMessage{MemberId: memberId, Page: page.Page, PageSize: page.PageSize}

	// 调用gRPC服务获取项目列表
	myProjectResponse, err := ProjectServiceClient.FindProjectByMemId(ctx, msg)
	if err != nil {
		// 解析gRPC错误并返回错误信息
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 如果返回的项目列表为空，初始化为空列表
	if myProjectResponse.Pm == nil {
		myProjectResponse.Pm = []*project.ProjectMessage{}
	}

	// 将gRPC返回的项目列表转换为ProjectAndMember对象
	var pms []*pro.ProjectAndMember
	copier.Copy(&pms, myProjectResponse.Pm)

	// 返回项目列表和总数量
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  pms,
		"total": myProjectResponse.Total,
	}))
}
