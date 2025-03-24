package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/pkg/model"
	"project-api/pkg/model/menu"
	"project-api/pkg/model/pro"
	common "project-common"
	"project-common/errs"
	"project-grpc/project"
	"strconv"
	"time"
)

type HandlerProject struct {
}

func New() *HandlerProject {
	return &HandlerProject{}
}

// Index 获取项目的菜单列表。
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
	menus := indexResponse.Menus
	var ms []*menu.Menu
	copier.Copy(&ms, menus)
	// 如果请求成功，返回成功响应，包含索引数据中的菜单信息。
	c.JSON(http.StatusOK, result.Success(ms))
}

// myProjectList 获取用户项目列表请求
// 该方法通过gRPC调用获取项目列表，并返回给客户端
func (p *HandlerProject) myProjectList(c *gin.Context) {
	// 初始化结果对象
	result := &common.Result{}

	// 1. 获取参数
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 从上下文中获取memberId
	memberIdStr, _ := c.Get("memberId")
	memberName := c.GetString("memberName")
	memberId := memberIdStr.(int64)

	// 绑定分页参数
	page := &model.Page{}
	page.Bind(c)
	selectBy := c.PostForm("selectBy")

	// 构建项目查询消息
	msg := &project.ProjectRpcMessage{
		MemberId:   memberId,
		MemberName: memberName,
		SelectBy:   selectBy,
		Page:       page.Page,
		PageSize:   page.PageSize,
	}

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

// projectTemplate 是一个处理获取项目模板信息的函数。
// 该函数通过 ProjectServiceClient 获取项目模板数据，并根据请求参数返回结果。
// 参数说明：
//   - c: gin.Context 类型，包含 HTTP 请求的相关信息和参数。
//
// 返回值：
//   - 通过 HTTP 响应返回项目模板列表（list）和总数量（total）。
func (p *HandlerProject) projectTemplate(c *gin.Context) {
	result := &common.Result{}

	// 创建上下文并设置超时时间，确保 RPC 调用不会无限期等待。
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 从 gin.Context 中提取用户相关的参数：memberId 和 memberName。
	memberId := c.GetInt64("memberId")
	memberName := c.GetString("memberName")

	// 绑定分页参数到 page 对象中。
	page := &model.Page{}
	page.Bind(c)

	// 获取视图类型参数并将其转换为整数。
	viewTypeStr := c.PostForm("viewType")
	viewType, _ := strconv.ParseInt(viewTypeStr, 10, 64)

	// 构造调用 RPC 的消息体，包含用户信息、分页信息和组织代码。
	msg := &project.ProjectRpcMessage{
		MemberId:         memberId,
		MemberName:       memberName,
		ViewType:         int32(viewType),
		Page:             page.Page,
		PageSize:         page.PageSize,
		OrganizationCode: c.GetString("organizationCode"),
	}

	// 调用远程服务获取项目模板数据。
	templateResponse, err := ProjectServiceClient.FindProjectTemplate(ctx, msg)
	if err != nil {
		// 如果发生错误，解析 gRPC 错误并返回失败响应。
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}

	// 将 RPC 响应中的项目模板数据复制到本地结构体中。
	var pms []*pro.ProjectTemplate
	copier.Copy(&pms, templateResponse.Ptm)
	if pms == nil {
		pms = []*pro.ProjectTemplate{}
	}

	// 遍历项目模板列表，确保 TaskStages 字段不为空。
	for _, v := range pms {
		if v.TaskStages == nil {
			// 如果 TaskStages 字段为空，将其设置为空列表。
			v.TaskStages = []*pro.TaskStagesOnlyName{}
		}
	}

	// 返回成功响应，包含项目模板列表和总数量。
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  pms, // 确保空值或 nil 转换为 []
		"total": templateResponse.Total,
	}))
}

// projectSave 保存项目信息
// 该方法接收一个 gin.Context 参数，从中提取出 memberId 和 organizationCode，
// 并结合用户请求的数据，调用项目保存服务，完成项目信息的保存或更新。
func (p *HandlerProject) projectSave(c *gin.Context) {
	result := &common.Result{}

	// 1. 获取参数
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	memberId := c.GetInt64("memberId")
	organizationCode := c.GetString("organizationCode")

	// 2. 解析请求体
	var req *pro.SaveProjectRequest
	c.ShouldBind(&req)

	// 3. 准备消息体，封装用户信息和项目信息
	msg := &project.ProjectRpcMessage{
		MemberId:         memberId,
		OrganizationCode: organizationCode,
		TemplateCode:     req.TemplateCode,
		Name:             req.Name,
		Id:               int64(req.Id),
		Description:      req.Description,
	}

	// 4. 调用项目保存服务
	saveProject, err := ProjectServiceClient.SaveProject(ctx, msg)
	if err != nil {
		// 处理保存失败的情况
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 5. 处理保存成功的情况
	var rsp *pro.SaveProject
	copier.Copy(&rsp, saveProject)
	c.JSON(http.StatusOK, result.Success(rsp))
}

// readProject 处理项目查询请求
// 该方法通过项目代码和成员ID获取项目详细信息，并将其返回给客户端
func (p *HandlerProject) readProject(c *gin.Context) {
	// 初始化结果对象，用于后续返回查询结果
	result := &common.Result{}

	// 从请求中获取项目代码和成员ID
	projectCode := c.PostForm("projectCode")
	memberId := c.GetInt64("memberId")

	// 创建一个带有超时的上下文，以确保请求不会无限期地等待
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel() // 确保在函数退出时取消上下文

	// 调用项目服务客户端的FindProjectDetail方法获取项目详细信息
	detail, err := ProjectServiceClient.FindProjectDetail(ctx, &project.ProjectRpcMessage{ProjectCode: projectCode, MemberId: memberId})
	if err != nil {
		// 如果发生错误，解析gRPC错误并返回错误信息
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 创建一个项目详情对象，并将获取的详情信息复制到该对象中
	pd := &pro.ProjectDetail{}
	copier.Copy(pd, detail)

	// 返回查询成功的响应，包含项目详情信息
	c.JSON(http.StatusOK, result.Success(pd))
}

// recycleProject 用于将项目标记为删除状态。
// 这个函数接收一个 gin.Context 参数，从中提取项目代码 projectCode，
// 并调用 gRPC 服务将项目标记为删除。如果操作成功，它将返回一个表示成功的响应；
// 如果失败，它将返回一个表示失败的响应。
func (p *HandlerProject) recycleProject(c *gin.Context) {
	// 初始化一个 Result 对象，用于后续构造响应。
	result := &common.Result{}

	// 从请求中获取项目代码 projectCode。
	projectCode := c.PostForm("projectCode")

	// 创建一个带有超时的上下文，以确保 gRPC 调用不会无限期地等待。
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel() // 确保在函数退出时取消上下文。

	// 调用 gRPC 服务，将项目标记为删除。
	_, err := ProjectServiceClient.UpdateDeletedProject(ctx, &project.ProjectRpcMessage{ProjectCode: projectCode, Deleted: true})
	if err != nil {
		// 如果发生错误，解析 gRPC 错误并返回失败的响应。
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 如果操作成功，返回成功的响应。
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// recoveryProject 是一个处理项目恢复请求的函数。
// 它接收一个 gin.Context 参数，从中提取项目代码 projectCode，
// 并调用 gRPC 服务更新项目的删除状态为未删除。
func (p *HandlerProject) recoveryProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ProjectServiceClient.UpdateDeletedProject(ctx, &project.ProjectRpcMessage{ProjectCode: projectCode, Deleted: false})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// collectProject 是一个处理项目收藏或取消收藏请求的函数。
// 它接收一个 gin.Context 参数，从中提取项目代码 projectCode、操作类型 type 和成员 ID memberId，
// 并调用 gRPC 服务更新项目的收藏状态。
func (p *HandlerProject) collectProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	collectType := c.PostForm("type")
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ProjectServiceClient.UpdateCollectProject(ctx, &project.ProjectRpcMessage{ProjectCode: projectCode, CollectType: collectType, MemberId: memberId})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// editProject 是一个处理编辑项目信息请求的函数。
// 它接收一个 gin.Context 参数，从中提取项目信息并绑定到 pro.ProjectReq 结构体，
// 并调用 gRPC 服务更新项目信息。
func (p *HandlerProject) editProject(c *gin.Context) {
	result := &common.Result{}
	var req *pro.ProjectReq
	_ = c.ShouldBind(&req)
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &project.UpdateProjectMessage{}
	copier.Copy(msg, req)
	msg.MemberId = memberId
	_, err := ProjectServiceClient.UpdateProject(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// getLogBySelfProject 是一个处理获取当前用户创建的项目日志的函数。
func (p *HandlerProject) getLogBySelfProject(c *gin.Context) {
	result := &common.Result{}
	var page = &model.Page{}
	page.Bind(c)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 创建一个 gRPC 消息，用于传递分页信息。
	msg := &project.ProjectRpcMessage{
		MemberId: c.GetInt64("memberId"),
		Page:     page.Page,
		PageSize: page.PageSize,
	}
	// 调用 gRPC 服务获取当前用户创建的项目日志。
	projectLogResponse, err := ProjectServiceClient.GetLogBySelfProject(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	// 将获取到的日志列表复制到结果中。
	var list []*model.ProjectLog
	copier.Copy(&list, projectLogResponse.List)
	if list == nil {
		list = []*model.ProjectLog{}
	}
	c.JSON(http.StatusOK, result.Success(list))
}

func (p *HandlerProject) nodeList(c *gin.Context) {
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	response, err := ProjectServiceClient.NodeList(ctx, &project.ProjectRpcMessage{})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var list []*model.ProjectNodeTree
	copier.Copy(&list, response.Nodes)
	c.JSON(http.StatusOK, result.Success(gin.H{
		"nodes": list,
	}))
}

func (p *HandlerProject) FindProjectByMemberId(memberId int64, projectCode string, taskCode string) (*pro.Project, bool, bool, *errs.BError) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &project.ProjectRpcMessage{
		MemberId:    memberId,
		ProjectCode: projectCode,
		TaskCode:    taskCode,
	}
	projectResponse, err := ProjectServiceClient.FindProjectByMemberId(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		return nil, false, false, errs.NewError(errs.ErrorCode(code), msg)
	}
	if projectResponse.Project == nil {
		return nil, false, false, nil
	}
	pr := &pro.Project{}
	copier.Copy(pr, projectResponse.Project)
	return pr, true, projectResponse.IsOwner, nil
}
