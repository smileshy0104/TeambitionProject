package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/pkg/model"
	"project-api/pkg/model/pro"
	"project-api/pkg/model/tasks"
	common "project-common"
	"project-common/errs"
	"project-grpc/task"
	"time"
)

type HandlerTask struct {
}

func NewTask() *HandlerTask {
	return &HandlerTask{}
}

// taskStages 处理任务阶段的请求。
// 该函数从gin.Context中提取参数，调用gRPC服务获取任务阶段信息，并返回给客户端。
func (t *HandlerTask) taskStages(c *gin.Context) {
	// 初始化结果对象
	result := &common.Result{}

	// 创建一个带有超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. 获取参数并校验参数的合法性
	projectCode := c.PostForm("projectCode")
	page := &model.Page{}
	page.Bind(c)

	// 2. 调用gRPC服务
	msg := &task.TaskReqMessage{
		MemberId:    c.GetInt64("memberId"),
		ProjectCode: projectCode,
		Page:        page.Page,
		PageSize:    page.PageSize,
	}
	stages, err := TaskServiceClient.TaskStages(ctx, msg)
	if err != nil {
		// 解析gRPC错误并返回响应
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 3. 处理响应
	var list []*tasks.TaskStagesResp
	copier.Copy(&list, stages.List)
	if list == nil {
		list = []*tasks.TaskStagesResp{}
	}
	for _, v := range list {
		v.TasksLoading = true  // 设置任务加载状态
		v.FixedCreator = false // 设置添加任务按钮定位
		v.ShowTaskCard = false // 设置是否显示创建卡片
		v.Tasks = []int{}
		v.DoneTasks = []int{}
		v.UnDoneTasks = []int{}
	}
	// 返回成功响应
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  list,
		"total": stages.Total,
		"page":  page.Page,
	}))
}

// memberProjectList 处理成员项目列表请求。
// 该函数接收一个 gin.Context 参数，从中提取项目代码和分页信息，
// 然后调用 gRPC 服务获取成员参与的项目列表，并将结果返回给客户端。
func (t *HandlerTask) memberProjectList(c *gin.Context) {
	// 初始化结果对象
	result := &common.Result{}

	// 创建一个带有超时的上下文，以确保请求不会无限期地等待
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. 获取参数并校验参数的合法性
	projectCode := c.PostForm("projectCode")
	page := &model.Page{}
	page.Bind(c)

	// 2. 调用 gRPC 服务
	msg := &task.TaskReqMessage{
		MemberId:    c.GetInt64("memberId"),
		ProjectCode: projectCode,
		Page:        page.Page,
		PageSize:    page.PageSize,
	}
	resp, err := TaskServiceClient.MemberProjectList(ctx, msg)
	if err != nil {
		// 解析 gRPC 错误并返回错误信息
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 初始化列表变量并复制从响应中获取的数据
	var list []*pro.MemberProjectResp
	copier.Copy(&list, resp.List)
	// 确保列表不为空
	if list == nil {
		list = []*pro.MemberProjectResp{}
	}

	// 返回成功响应，包含项目列表和分页信息
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  list,
		"total": resp.Total,
		"page":  page.Page,
	}))
}

// taskList 处理任务列表的请求。
// 该函数从请求中提取阶段代码(stageCode)，并调用TaskServiceClient的服务获取任务列表。
// 它还负责处理可能发生的错误，并将结果以JSON格式返回给客户端。
func (t *HandlerTask) taskList(c *gin.Context) {
	// 初始化结果对象，用于存储处理结果。
	result := &common.Result{}

	// 从请求中获取阶段代码(stageCode)。
	stageCode := c.PostForm("stageCode")

	// 创建一个带有超时的上下文，以确保请求不会无限期地等待。
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 调用TaskServiceClient的服务获取任务列表。
	list, err := TaskServiceClient.TaskList(ctx, &task.TaskReqMessage{StageCode: stageCode, MemberId: c.GetInt64("memberId")})
	if err != nil {
		// 如果发生错误，解析gRPC错误并返回错误信息。
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 初始化任务显示列表。
	var taskDisplayList []*tasks.TaskDisplay
	// 将获取的任务列表复制到任务显示列表中。
	copier.Copy(&taskDisplayList, list.List)
	// 确保任务显示列表不为空。
	if taskDisplayList == nil {
		taskDisplayList = []*tasks.TaskDisplay{}
	}

	// 为确保返回给前端的数据中没有null值，对任务显示列表中的每个任务进行检查和初始化可能为空的字段。
	for _, v := range taskDisplayList {
		// 如果任务的标签为空，则初始化为空数组。
		if v.Tags == nil {
			v.Tags = []int{}
		}
		// 如果任务的子任务计数为空，则初始化为空数组。
		if v.ChildCount == nil {
			v.ChildCount = []int{}
		}
	}

	// 返回成功结果和任务显示列表。
	c.JSON(http.StatusOK, result.Success(taskDisplayList))
}

// saveTask 保存任务信息。
// 该方法首先绑定请求参数，然后通过gRPC调用任务保存服务。
// 参数: c *gin.Context - Gin框架的上下文，用于处理HTTP请求和响应。
func (t *HandlerTask) saveTask(c *gin.Context) {
	result := &common.Result{}
	var req *tasks.TaskSaveReq

	// 绑定请求参数到req对象。
	c.ShouldBind(&req)

	// 创建一个带有2秒超时的上下文，用于控制gRPC调用的最长执行时间。
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 构建任务保存请求消息。
	msg := &task.TaskReqMessage{
		ProjectCode: req.ProjectCode,
		Name:        req.Name,
		StageCode:   req.StageCode,
		AssignTo:    req.AssignTo,
		MemberId:    c.GetInt64("memberId"),
	}

	// 调用gRPC服务保存任务，并处理错误。
	taskMessage, err := TaskServiceClient.SaveTask(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 创建一个任务显示对象，并将gRPC响应数据复制到该对象中。
	td := &tasks.TaskDisplay{}
	copier.Copy(td, taskMessage)

	// 确保任务显示对象中的Tags和ChildCount字段不为空。
	if td != nil {
		if td.Tags == nil {
			td.Tags = []int{}
		}
		if td.ChildCount == nil {
			td.ChildCount = []int{}
		}
	}

	// 返回保存成功的结果。
	c.JSON(http.StatusOK, result.Success(td))
}

// taskSort 处理任务排序请求。
// 该函数接收一个 gin.Context 参数，从中解析任务排序请求信息，并调用 TaskServiceClient 的 TaskSort 方法进行处理。
// 如果处理成功，返回成功结果；如果处理失败，解析错误并返回错误信息。
func (t *HandlerTask) taskSort(c *gin.Context) {
	result := &common.Result{}
	var req *tasks.TaskSortReq
	c.ShouldBind(&req)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		PreTaskCode:  req.PreTaskCode,
		NextTaskCode: req.NextTaskCode,
		ToStageCode:  req.ToStageCode,
	}
	_, err := TaskServiceClient.TaskSort(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// myTaskList 获取用户的任务列表。
// 该函数根据用户ID和任务类型等参数，调用 TaskServiceClient 的 MyTaskList 方法获取任务列表。
// 如果获取成功，整理任务列表信息并返回；如果获取失败，解析错误并返回错误信息。
func (t *HandlerTask) myTaskList(c *gin.Context) {
	result := &common.Result{}
	var req *tasks.MyTaskReq
	c.ShouldBind(&req)
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		MemberId: memberId,
		TaskType: int32(req.TaskType),
		Type:     int32(req.Type),
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	myTaskListResponse, err := TaskServiceClient.MyTaskList(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var myTaskList []*tasks.MyTaskDisplay
	copier.Copy(&myTaskList, myTaskListResponse.List)
	if myTaskList == nil {
		myTaskList = []*tasks.MyTaskDisplay{}
	}
	for _, v := range myTaskList {
		v.ProjectInfo = tasks.ProjectInfo{
			Name: v.ProjectName,
			Code: v.ProjectCode,
		}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  myTaskList,
		"total": myTaskListResponse.Total,
	}))
}
