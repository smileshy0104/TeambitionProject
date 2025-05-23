package project

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"path"
	"project-api/pkg/model"
	"project-api/pkg/model/pro"
	"project-api/pkg/model/tasks"
	common "project-common"
	"project-common/errs"
	"project-common/minio"
	"project-common/tms"
	"project-grpc/task"
	"strconv"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

// editTask 修改任务信息。
// 该方法首先绑定请求参数，然后通过gRPC调用任务保存服务。
// 参数: c *gin.Context - Gin框架的上下文，用于处理HTTP请求和响应。
func (t *HandlerTask) editTask(c *gin.Context) {
	result := &common.Result{}
	var req *tasks.TaskEditReq

	// 绑定请求参数到req对象。
	c.ShouldBind(&req)

	fmt.Println(req)
	// 创建一个带有2秒超时的上下文，用于控制gRPC调用的最长执行时间。
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构建任务保存请求消息。
	msg := &task.TaskReqMessage{
		Name:     req.Name,
		TaskCode: req.TaskCode,
		AssignTo: req.AssignTo,
		MemberId: c.GetInt64("memberId"),
	}

	// 调用gRPC服务保存任务，并处理错误。
	_, err := TaskServiceClient.EditTask(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 返回保存成功的结果。
	c.JSON(http.StatusOK, result.Success("修改成功！"))
}

// taskSort 处理任务排序请求。
// 该函数接收一个 gin.Context 参数，从中解析任务排序请求信息，并调用 TaskServiceClient 的 TaskSort 方法进行处理。
// 如果处理成功，返回成功结果；如果处理失败，解析错误并返回错误信息。
func (t *HandlerTask) taskSort(c *gin.Context) {
	result := &common.Result{}
	var req *tasks.TaskSortReq
	c.ShouldBind(&req)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

// readTask 读取任务详情。
func (t *HandlerTask) readTask(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		TaskCode: taskCode,
		MemberId: c.GetInt64("memberId"),
	}
	// 调用 TaskServiceClient 的 ReadTask 方法获取任务详情。
	taskMessage, err := TaskServiceClient.ReadTask(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	// 将任务详情复制到任务显示对象中。
	td := &tasks.TaskDisplay{}
	copier.Copy(td, taskMessage)
	if td != nil {
		if td.Tags == nil {
			td.Tags = []int{}
		}
		if td.ChildCount == nil {
			td.ChildCount = []int{}
		}
	}
	c.JSON(200, result.Success(td))
}

// listTaskMember 获取任务成员列表。
func (t *HandlerTask) listTaskMember(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	page := &model.Page{}
	page.Bind(c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		TaskCode: taskCode,
		MemberId: c.GetInt64("memberId"),
		Page:     page.Page,
		PageSize: page.PageSize,
	}
	// 调用 TaskServiceClient 的 ListTaskMember 方法获取任务成员列表。
	taskMemberResponse, err := TaskServiceClient.ListTaskMember(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	// 将任务成员列表复制到任务成员显示对象中。
	var tms []*tasks.TaskMember
	copier.Copy(&tms, taskMemberResponse.List)
	if tms == nil {
		tms = []*tasks.TaskMember{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  tms,
		"total": taskMemberResponse.Total,
		"page":  page.Page,
	}))
}

// taskLog 获取任务日志列表。
func (t *HandlerTask) taskLog(c *gin.Context) {
	result := &common.Result{}
	var req *model.TaskLogReq
	c.ShouldBind(&req)
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 创建一个任务日志请求消息对象，并设置相关参数。
	msg := &task.TaskReqMessage{
		TaskCode: req.TaskCode,
		MemberId: c.GetInt64("memberId"),
		Page:     int64(req.Page),
		PageSize: int64(req.PageSize),
		All:      int32(req.All),
		Comment:  int32(req.Comment),
	}
	// 调用 TaskServiceClient 的 TaskLog 方法获取任务日志列表。
	taskLogResponse, err := TaskServiceClient.TaskLog(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var tms []*model.ProjectLogDisplay
	copier.Copy(&tms, taskLogResponse.List)
	if tms == nil {
		tms = []*model.ProjectLogDisplay{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  tms,
		"total": taskLogResponse.Total,
		"page":  req.Page,
	}))
}

// taskWorkTimeList 获取任务工时列表。
func (t *HandlerTask) taskWorkTimeList(c *gin.Context) {
	taskCode := c.PostForm("taskCode")
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		TaskCode: taskCode,
		MemberId: c.GetInt64("memberId"),
	}
	// 调用 TaskServiceClient 的 TaskWorkTimeList 方法获取任务工时列表。
	taskWorkTimeResponse, err := TaskServiceClient.TaskWorkTimeList(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	// 将任务工时列表复制到任务工时显示对象中。
	var tms []*model.TaskWorkTime
	copier.Copy(&tms, taskWorkTimeResponse.List)
	if tms == nil {
		tms = []*model.TaskWorkTime{}
	}
	c.JSON(http.StatusOK, result.Success(tms))
}

// saveTaskWorkTime 保存任务工时。
func (t *HandlerTask) saveTaskWorkTime(c *gin.Context) {
	result := &common.Result{}
	var req *model.SaveTaskWorkTimeReq
	c.ShouldBind(&req)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 创建一个任务工时请求消息对象，并设置相关参数。
	msg := &task.TaskReqMessage{
		TaskCode:  req.TaskCode,
		MemberId:  c.GetInt64("memberId"),
		Content:   req.Content,
		Num:       int32(req.Num),
		BeginTime: tms.ParseTime(req.BeginTime),
	}
	// 调用 TaskServiceClient 的 SaveTaskWorkTime 方法保存任务工时。
	_, err := TaskServiceClient.SaveTaskWorkTime(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// uploadFiles 上传文件。
func (t *HandlerTask) uploadFiles(c *gin.Context) {
	result := &common.Result{}
	// 获取上传文件的参数。
	req := model.UploadFileReq{}
	c.ShouldBind(&req)
	//处理文件
	multipartForm, _ := c.MultipartForm()
	file := multipartForm.File
	//假设只上传一个文件
	uploadFile := file["file"][0]
	//key := ""
	// 第一种 没有达成分片的条件    判断是否是单片上传
	// TODO 上传文件到minio
	key := "msproject/" + req.Filename
	minioClient, err := minio.New(
		"localhost:9009",
		"XPk01wTsdntPtZmRPMe4",
		"B0rProcvGvGOL68fWI9xY0E5GXn9K8q5iz61XPSI",
		false)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(-999, err.Error()))
		return
	}
	bucketName := "msproject"
	if req.TotalChunks == 1 {
		// TODO 上传本地目录文件
		////不分片
		//path := "upload/" + req.ProjectCode + "/" + req.TaskCode + "/" + tms.FormatYMD(time.Now())
		//if !fs.IsExist(path) {
		//	os.MkdirAll(path, os.ModePerm)
		//}
		//dst := path + "/" + req.Filename
		//key = dst
		//// 保存上传的文件到指定路径。
		//err := c.SaveUploadedFile(uploadFile, dst)
		//if err != nil {
		//	c.JSON(http.StatusOK, result.Fail(-999, err.Error()))
		//	return
		//}

		// TODO 上传到minio
		// 打开文件，如果文件不存在则创建文件。
		open, err := uploadFile.Open()
		if err != nil {
			c.JSON(http.StatusOK, result.Fail(-999, err.Error()))
			return
		}
		defer open.Close()
		// 读取文件内容并上传到 MinIO 服务器。
		buf := make([]byte, req.CurrentChunkSize)
		open.Read(buf)
		// 上传文件到 MinIO 服务器。
		_, err = minioClient.Put(
			context.Background(),
			bucketName,
			req.Filename,
			buf,
			int64(req.TotalSize),
			uploadFile.Header.Get("Content-Type"),
		)
		if err != nil {
			c.JSON(http.StatusOK, result.Fail(-999, err.Error()))
			return
		}
	}
	// 第二种 是分片上传
	if req.TotalChunks > 1 {
		// TODO 分片上传至本地目录文件
		////分片上传 无非就是先把每次的存储起来 追加就可以了
		//path := "upload/" + req.ProjectCode + "/" + req.TaskCode + "/" + tms.FormatYMD(time.Now())
		//if !fs.IsExist(path) {
		//	os.MkdirAll(path, os.ModePerm)
		//}
		//fileName := path + "/" + req.Identifier
		//// 打开文件，如果文件不存在则创建文件。
		//openFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
		//if err != nil {
		//	c.JSON(http.StatusOK, result.Fail(-999, err.Error()))
		//	return
		//}
		//// 打开上传的文件。
		//open, err := uploadFile.Open()
		//if err != nil {
		//	c.JSON(http.StatusOK, result.Fail(-999, err.Error()))
		//	return
		//}
		//defer open.Close()
		//// 从上传的文件中读取数据并写入文件。
		//buf := make([]byte, req.CurrentChunkSize)
		//open.Read(buf)
		//openFile.Write(buf)
		//openFile.Close()
		//key = fileName
		//// 如果是最后一个分片，则重命名文件。
		//if req.TotalChunks == req.ChunkNumber {
		//	//最后一个分片了
		//	newPath := path + "/" + req.Filename
		//	key = newPath
		//	os.Rename(fileName, newPath)
		//}

		// TODO 分片上传至minio
		//分片上传 无非就是先把每次的存储起来 追加就可以了
		open, err := uploadFile.Open()
		if err != nil {
			c.JSON(http.StatusOK, result.Fail(-999, err.Error()))
			return
		}
		defer open.Close()
		buf := make([]byte, req.CurrentChunkSize)
		// 读取文件内容并上传到 MinIO 服务器。
		open.Read(buf)
		formatInt := strconv.FormatInt(int64(req.ChunkNumber), 10)
		// 上传文件到 MinIO 服务器。
		_, err = minioClient.Put(
			context.Background(),
			bucketName,
			req.Filename+"_"+formatInt,
			buf,
			int64(req.CurrentChunkSize),
			uploadFile.Header.Get("Content-Type"),
		)
		if err != nil {
			c.JSON(http.StatusOK, result.Fail(-999, err.Error()))
			return
		}
		// 如果是最后一个分片，则合并分片。
		if req.TotalChunks == req.ChunkNumber {
			//最后一个分片了 合并
			_, err := minioClient.Compose(context.Background(), bucketName, req.Filename, req.TotalChunks)
			if err != nil {
				c.JSON(http.StatusOK, result.Fail(-999, err.Error()))
				return
			}
		}
	}
	//调用服务 存入file表
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 创建一个任务文件请求消息对象，并设置相关参数。
	fileUrl := "http://localhost:9009/" + key
	msg := &task.TaskFileReqMessage{
		TaskCode:         req.TaskCode,
		ProjectCode:      req.ProjectCode,
		OrganizationCode: c.GetString("organizationCode"),
		PathName:         key,
		FileName:         req.Filename,
		Size:             int64(req.TotalSize),
		Extension:        path.Ext(key),
		FileUrl:          fileUrl,
		FileType:         file["file"][0].Header.Get("Content-Type"),
		MemberId:         c.GetInt64("memberId"),
	}
	// 如果是最后一个分片，则重命名文件。
	// 调用 TaskServiceClient 的 SaveTaskFile 方法保存任务文件。
	if req.TotalChunks == req.ChunkNumber {
		_, err := TaskServiceClient.SaveTaskFile(ctx, msg)
		if err != nil {
			code, msg := errs.ParseGrpcError(err)
			c.JSON(http.StatusOK, result.Fail(code, msg))
		}
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"file":        key,
		"hash":        "",
		"key":         key,
		"url":         "http://localhost:9009/" + key,
		"projectName": req.ProjectName,
	}))
}

// taskSources 获取任务来源。
func (t *HandlerTask) taskSources(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 调用 TaskServiceClient 的 TaskSources 方法获取任务来源。
	sources, err := TaskServiceClient.TaskSources(ctx, &task.TaskReqMessage{TaskCode: taskCode})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var slList []*model.SourceLink
	copier.Copy(&slList, sources.List)
	if slList == nil {
		slList = []*model.SourceLink{}
	}
	c.JSON(http.StatusOK, result.Success(slList))
}

// createComment 创建评论。
func (t *HandlerTask) createComment(c *gin.Context) {
	result := &common.Result{}
	// 获取评论参数。
	req := model.CommentReq{}
	c.ShouldBind(&req)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		TaskCode:       req.TaskCode,
		CommentContent: req.Comment,
		Mentions:       req.Mentions,
		MemberId:       c.GetInt64("memberId"),
	}
	// 调用 TaskServiceClient 的 CreateComment 方法创建评论。
	_, err := TaskServiceClient.CreateComment(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	c.JSON(http.StatusOK, result.Success(true))
}
