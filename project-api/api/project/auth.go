package project

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/pkg/model"
	common "project-common"
	"project-common/errs"
	"project-grpc/auth"
	"time"
)

type HandlerAuth struct {
}

// authList 处理授权列表的请求。
// 该函数从上下文中获取组织代码，绑定分页信息，并调用远程服务获取授权信息列表。
func (a *HandlerAuth) authList(c *gin.Context) {
	// 初始化结果对象，用于后续的响应。
	result := &common.Result{}

	// 从上下文中获取组织代码。
	organizationCode := c.GetString("organizationCode")

	// 初始化并绑定分页对象。
	var page = &model.Page{}
	page.Bind(c)

	// 创建一个带有超时的上下文，以确保请求不会无限期地等待。
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 构造授权请求消息。
	msg := &auth.AuthReqMessage{
		OrganizationCode: organizationCode,
		Page:             page.Page,
		PageSize:         page.PageSize,
	}

	// 调用 AuthServiceClient 的 AuthList 方法获取授权列表。
	response, err := AuthServiceClient.AuthList(ctx, msg)
	if err != nil {
		// 解析gRPC错误，返回失败的响应。
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 初始化授权列表，并将响应数据复制到列表中。
	var authList []*model.ProjectAuth
	copier.Copy(&authList, response.List)

	// 确保列表不为空。
	if authList == nil {
		authList = []*model.ProjectAuth{}
	}

	// 返回成功的响应，包含总记录数、列表数据和当前页码。
	c.JSON(http.StatusOK, result.Success(gin.H{
		"total": response.Total,
		"list":  authList,
		"page":  page.Page,
	}))
}

// apply 函数处理项目权限申请。
// 它接收一个 gin.Context 参数，从中解析项目权限请求信息，
// 并调用 AuthServiceClient 的 Apply 方法处理权限申请。
// 处理结果以 JSON 格式返回给客户端。
func (a *HandlerAuth) apply(c *gin.Context) {
	// 初始化结果对象，用于后续返回结果。
	result := &common.Result{}

	// 解析请求体，获取项目权限请求信息。
	var req *model.ProjectAuthReq
	c.ShouldBind(&req)

	// 初始化节点列表，用于存储解析后的节点信息。
	var nodes []string
	// 如果请求中的节点信息不为空，则解析 JSON 格式的节点信息。
	if req.Nodes != "" {
		json.Unmarshal([]byte(req.Nodes), &nodes)
	}

	// 创建一个带有超时的上下文，用于控制后续操作的执行时间。
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 构建权限请求消息。
	msg := &auth.AuthReqMessage{
		Action: req.Action,
		AuthId: req.Id,
		Nodes:  nodes,
	}

	// 调用 AuthServiceClient 的 Apply 方法处理权限申请。
	applyResponse, err := AuthServiceClient.Apply(ctx, msg)
	if err != nil {
		// 如果发生错误，解析 gRPC 错误并返回错误信息。
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 将 Apply 方法返回的列表数据转换为 ProjectNodeAuthTree 类型的列表。
	var list []*model.ProjectNodeAuthTree
	copier.Copy(&list, applyResponse.List)

	// 将 Apply 方法返回的选中节点列表转换为字符串列表。
	var checkedList []string
	copier.Copy(&checkedList, applyResponse.CheckedList)

	// 返回成功结果，包含处理后的列表数据和选中节点列表。
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":        list,
		"checkedList": checkedList,
	}))
}

// GetAuthNodes 获取成员的授权节点
// 该方法从HTTP上下文中提取成员ID，创建一个授权请求消息，
// 并在带有超时的上下文中调用gRPC服务以获取授权节点列表。
func (a *HandlerAuth) GetAuthNodes(c *gin.Context) ([]string, error) {
	// 从HTTP上下文中获取成员ID
	memberId := c.GetInt64("memberId")

	// 创建授权请求消息，包含成员ID
	msg := &auth.AuthReqMessage{
		MemberId: memberId,
	}

	// 创建一个带有2秒超时的上下文，以防止长时间运行的请求
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel() // 确保在函数退出时取消上下文

	// 调用gRPC服务，获取成员的授权节点列表
	response, err := AuthServiceClient.AuthNodesByMemberId(ctx, msg)
	if err != nil {
		// 解析gRPC错误，获取错误代码和消息
		code, msg := errs.ParseGrpcError(err)
		// 创建并返回自定义错误
		return nil, errs.NewError(errs.ErrorCode(code), msg)
	}

	// 返回授权节点列表
	return response.List, err
}

func NewAuth() *HandlerAuth {
	return &HandlerAuth{}
}
