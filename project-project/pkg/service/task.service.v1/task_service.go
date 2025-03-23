package task_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"project-common/encrypts"
	"project-common/errs"
	"project-common/tms"
	"project-grpc/task"
	"project-grpc/user/login"
	"project-project/internal/dao"
	dataNew "project-project/internal/data"
	"project-project/internal/data/pro"
	data "project-project/internal/data/task"
	"project-project/internal/database"
	"project-project/internal/database/tran"
	"project-project/internal/repo"
	"project-project/internal/rpc"
	"project-project/pkg/model"
	"time"
)

// TaskService 提供了任务相关的服务功能，继承了 task.UnimplementedTaskServiceServer。
// 它通过多个数据存储库来管理缓存、事务以及项目、任务等数据。
type TaskService struct {
	task.UnimplementedTaskServiceServer                             // 继承自任务服务的未实现服务服务器
	cache                               repo.Cache                  // 缓存接口，用于快速数据访问
	transaction                         tran.Transaction            // 事务接口，用于处理数据的一致性和完整性
	projectRepo                         repo.ProjectRepo            // 项目数据存储库，管理项目相关数据
	projectTemplateRepo                 repo.ProjectTemplateRepo    // 项目模板数据存储库，管理项目模板数据
	taskStagesTemplateRepo              repo.TaskStagesTemplateRepo // 任务阶段模板数据存储库，管理任务阶段模板数据
	taskStagesRepo                      repo.TaskStagesRepo         // 任务阶段数据存储库，管理任务阶段数据
	taskRepo                            repo.TaskRepo               // 任务数据存储库，管理任务数据
	projectLogRepo                      repo.ProjectLogRepo
	taskWorkTimeRepo                    repo.TaskWorkTimeRepo
	fileRepo                            repo.FileRepo
	sourceLinkRepo                      repo.SourceLinkRepo
}

// New 创建并返回一个新的 TaskService 实例。
// 它初始化了 TaskService 结构体，并注入了必要的数据存储库实例。
func New() *TaskService {
	return &TaskService{
		cache:                  dao.Rc,                         // 使用预定义的缓存实例
		transaction:            dao.NewTransaction(),           // 创建新的事务实例
		projectRepo:            dao.NewProjectDao(),            // 创建新的项目数据存储库实例
		projectTemplateRepo:    dao.NewProjectTemplateDao(),    // 创建新的项目模板数据存储库实例
		taskStagesTemplateRepo: dao.NewTaskStagesTemplateDao(), // 创建新的任务阶段模板数据存储库实例
		taskStagesRepo:         dao.NewTaskStagesDao(),         // 创建新的任务阶段数据存储库实例
		taskRepo:               dao.NewTaskDao(),               // 创建新的任务数据存储库实例
		projectLogRepo:         dao.NewProjectLogDao(),
		taskWorkTimeRepo:       dao.NewTaskWorkTimeDao(),
		fileRepo:               dao.NewFileDao(),
		sourceLinkRepo:         dao.NewSourceLinkDao(),
	}
}

// TaskStages 获取任务阶段信息
// 该方法根据项目代码、页码和页面大小获取任务阶段信息
func (t *TaskService) TaskStages(co context.Context, msg *task.TaskReqMessage) (*task.TaskStagesResponse, error) {
	// 解密项目代码
	projectCode := encrypts.DecryptNoErr(msg.ProjectCode)
	// 获取页码和页面大小
	page := msg.Page
	pageSize := msg.PageSize
	// 创建一个带有超时的上下文，以防止操作无限期阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 调用存储库方法获取任务阶段信息
	stages, total, err := t.taskStagesRepo.FindStagesByProjectId(ctx, projectCode, page, pageSize)
	if err != nil {
		// 如果发生错误，记录错误信息并返回通用的 gRPC 错误
		zap.L().Error("project SaveProject taskStagesRepo.FindStagesByProjectId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}

	// 初始化任务阶段消息列表并进行数据复制
	var tsMessages []*task.TaskStagesMessage
	copier.Copy(&tsMessages, stages)
	// 如果列表为空，直接返回空列表和总数 0
	if tsMessages == nil {
		return &task.TaskStagesResponse{List: tsMessages, Total: 0}, nil
	}

	// 将阶段信息转换为映射格式
	stagesMap := data.ToTaskStagesMap(stages)
	// 遍历任务阶段消息列表，填充详细信息
	for _, v := range tsMessages {
		taskStages := stagesMap[int(v.Id)]
		// 加密任务阶段 ID
		v.Code = encrypts.EncryptNoErr(int64(v.Id))
		// 格式化创建时间
		v.CreateTime = tms.FormatByMill(taskStages.CreateTime)
		// 设置项目代码
		v.ProjectCode = msg.ProjectCode
	}

	// 返回包含任务阶段列表和总数的响应对象
	return &task.TaskStagesResponse{List: tsMessages, Total: total}, nil
}

// MemberProjectList 查询项目成员列表
// 该方法首先根据项目代码查询项目成员信息，然后根据成员ID请求用户信息，
// 最后组装并返回项目成员的详细信息列表。
func (t *TaskService) MemberProjectList(co context.Context, msg *task.TaskReqMessage) (*task.MemberProjectResponse, error) {
	// 1. 去 project_member表 去查询 用户id列表
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	projectCode := encrypts.DecryptNoErr(msg.ProjectCode)
	// 查询项目成员信息通过项目code
	projectMembers, total, err := t.projectRepo.FindProjectMemberByPid(ctx, projectCode)
	if err != nil {
		zap.L().Error("project MemberProjectList projectRepo.FindProjectMemberByPid error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	// 2.拿上用户id列表 去请求用户信息
	if projectMembers == nil || len(projectMembers) <= 0 {
		return &task.MemberProjectResponse{List: nil, Total: 0}, nil
	}
	var mIds []int64
	// 创建一个映射，键为成员ID，值为项目成员对象
	pmMap := make(map[int64]*pro.ProjectMember)
	// 遍历项目成员列表，将成员ID添加到 mIds 切片并创建一个映射，键为成员ID，值为项目成员对象
	for _, v := range projectMembers {
		mIds = append(mIds, v.MemberCode)
		pmMap[v.MemberCode] = v
	}
	// 请求用户信息
	userMsg := &login.UserMessage{
		MIds: mIds,
	}
	// 调用 RPC 客户端请求用户信息
	memberMessageList, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, userMsg)
	if err != nil {
		zap.L().Error("project MemberProjectList LoginServiceClient.FindMemInfoByIds error", zap.Error(err))
		return nil, err
	}
	var list []*task.MemberProjectMessage
	// 遍历用户信息列表，将用户信息转换为 MemberProjectMessage 并添加到列表中
	for _, v := range memberMessageList.List {
		owner := pmMap[v.Id].IsOwner
		mpm := &task.MemberProjectMessage{
			MemberCode: v.Id,
			Name:       v.Name,
			Avatar:     v.Avatar,
			Email:      v.Email,
			Code:       v.Code,
		}
		if v.Id == owner {
			mpm.IsOwner = 1
		}
		list = append(list, mpm)
	}
	return &task.MemberProjectResponse{List: list, Total: total}, nil
}

// TaskList 查询任务列表
// 该方法根据阶段代码查询任务，并处理任务的隐私设置及成员信息
func (t *TaskService) TaskList(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskListResponse, error) {
	// 解密阶段代码
	stageCode := encrypts.DecryptNoErr(msg.StageCode)
	// 创建一个带有超时的上下文，以防止长时间运行的任务导致服务阻塞
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 根据阶段代码查询任务列表
	taskList, err := t.taskRepo.FindTaskByStageCode(c, int(stageCode))
	if err != nil {
		// 记录错误日志，并返回数据库错误
		zap.L().Error("project task TaskList FindTaskByStageCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	// 初始化任务显示列表和成员ID列表
	var taskDisplayList []*data.TaskDisplay
	var mIds []int64
	// 遍历任务列表，处理每个任务
	for _, v := range taskList {
		// 将任务转换为显示格式
		display := v.ToTaskDisplay()
		// 如果任务是私有的，检查当前用户是否有权限查看
		if v.Private == 1 {
			// 查询任务成员信息
			taskMember, err := t.taskRepo.FindTaskMemberByTaskId(ctx, v.Id, msg.MemberId)
			if err != nil {
				// 记录错误日志，并返回数据库错误
				zap.L().Error("project task TaskList taskRepo.FindTaskMemberByTaskId error", zap.Error(err))
				return nil, errs.GrpcError(model.DBError)
			}
			// 根据查询结果设置任务的可读状态
			if taskMember != nil {
				display.CanRead = model.CanRead
			} else {
				display.CanRead = model.NoCanRead
			}
		}
		// 将处理后的任务添加到显示列表中
		taskDisplayList = append(taskDisplayList, display)
		// 添加任务分配给的成员ID到成员ID列表中
		mIds = append(mIds, v.AssignTo)
	}
	// 如果成员ID列表为空，直接返回空的列表响应
	if mIds == nil || len(mIds) <= 0 {
		return &task.TaskListResponse{List: nil}, nil
	}
	// 查询成员信息
	messageList, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, &login.UserMessage{MIds: mIds})
	if err != nil {
		// 记录错误日志，并返回查询成员信息时遇到的错误
		zap.L().Error("project task TaskList LoginServiceClient.FindMemInfoByIds error", zap.Error(err))
		return nil, err
	}
	// 将查询到的成员信息存储到映射中，以便后续快速查找
	memberMap := make(map[int64]*login.MemberMessage)
	for _, v := range messageList.List {
		memberMap[v.Id] = v
	}
	// 为每个任务设置执行人信息
	for _, v := range taskDisplayList {
		// 从成员信息映射中获取执行人信息
		message := memberMap[encrypts.DecryptNoErr(v.AssignTo)]
		e := data.Executor{
			Name:   message.Name,
			Avatar: message.Avatar,
		}
		v.Executor = e
	}
	// 初始化任务消息列表，并将任务显示列表复制到任务消息列表中
	var taskMessageList []*task.TaskMessage
	copier.Copy(&taskMessageList, taskDisplayList)
	// 返回任务列表响应
	return &task.TaskListResponse{List: taskMessageList}, nil
}

// SaveTask 保存任务信息，包括任务的基本信息、分配信息等。
func (t *TaskService) SaveTask(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskMessage, error) {
	//1. 检查业务逻辑
	if msg.Name == "" {
		return nil, errs.GrpcError(model.TaskNameNotNull)
	}
	stageCode := encrypts.DecryptNoErr(msg.StageCode)
	taskStages, err := t.taskStagesRepo.FindById(ctx, int(stageCode))
	if err != nil {
		zap.L().Error("project task SaveTask taskStagesRepo.FindById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if taskStages == nil {
		return nil, errs.GrpcError(model.TaskStagesNotNull)
	}
	projectCode := encrypts.DecryptNoErr(msg.ProjectCode)
	// 查询项目信息
	project, err := t.projectRepo.FindProjectById(ctx, projectCode)
	if err != nil {
		zap.L().Error("project task SaveTask projectRepo.FindProjectById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if project == nil || project.Deleted == model.Deleted {
		return nil, errs.GrpcError(model.ProjectAlreadyDeleted)
	}
	// 查询当前项目下的最大任务编号和最大任务排序号
	maxIdNum, err := t.taskRepo.FindTaskMaxIdNum(ctx, projectCode)
	if err != nil {
		zap.L().Error("project task SaveTask taskRepo.FindTaskMaxIdNum error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if maxIdNum == nil {
		a := 0
		maxIdNum = &a
	}
	// 查询当前阶段下的最大任务排序号
	maxSort, err := t.taskRepo.FindTaskSort(ctx, projectCode, stageCode)
	if err != nil {
		zap.L().Error("project task SaveTask taskRepo.FindTaskSort error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if maxSort == nil {
		a := 0
		maxSort = &a
	}
	assignTo := encrypts.DecryptNoErr(msg.AssignTo)
	ts := &data.Task{
		Name:        msg.Name,
		CreateTime:  time.Now().UnixMilli(),
		CreateBy:    msg.MemberId,
		AssignTo:    assignTo,
		ProjectCode: projectCode,
		StageCode:   int(stageCode),
		IdNum:       *maxIdNum + 1,
		Private:     project.OpenTaskPrivate,
		Sort:        *maxSort + 65536,
		BeginTime:   time.Now().UnixMilli(),
		EndTime:     time.Now().Add(2 * 24 * time.Hour).UnixMilli(),
	}
	err = t.transaction.Action(func(conn database.DbConn) error {
		err = t.taskRepo.SaveTask(ctx, conn, ts)
		if err != nil {
			zap.L().Error("project task SaveTask taskRepo.SaveTask error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}

		tm := &data.TaskMember{
			MemberCode: assignTo,
			TaskCode:   ts.Id,
			JoinTime:   time.Now().UnixMilli(),
			IsOwner:    model.Owner,
		}
		if assignTo == msg.MemberId {
			tm.IsExecutor = model.Executor
		}
		err = t.taskRepo.SaveTaskMember(ctx, conn, tm)
		if err != nil {
			zap.L().Error("project task SaveTask taskRepo.SaveTaskMember error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	display := ts.ToTaskDisplay()
	member, err := rpc.LoginServiceClient.FindMemInfoById(ctx, &login.UserMessage{MemId: assignTo})
	if err != nil {
		return nil, err
	}
	display.Executor = data.Executor{
		Name:   member.Name,
		Avatar: member.Avatar,
		Code:   member.Code,
	}
	tm := &task.TaskMessage{}
	copier.Copy(tm, display)
	return tm, nil
}

// TaskSort 负责处理任务排序请求。
// 该函数接收一个上下文和一个任务请求消息，然后根据消息中的信息对任务进行排序。
func (t *TaskService) TaskSort(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskSortResponse, error) {
	// 解密前置任务代码和目标阶段代码，以获取实际的任务和阶段标识。
	preTaskCode := encrypts.DecryptNoErr(msg.PreTaskCode)
	toStageCode := encrypts.DecryptNoErr(msg.ToStageCode)

	// 如果前置任务代码和下一任务代码相同，则无需进行任何操作，直接返回空响应。
	if msg.PreTaskCode == msg.NextTaskCode {
		return &task.TaskSortResponse{}, nil
	}

	// 调用sortTask方法进行任务排序，如果排序失败，则返回相应的错误。
	err := t.sortTask(preTaskCode, msg.NextTaskCode, toStageCode)
	if err != nil {
		return nil, err
	}

	// 如果排序成功，返回空响应对象。
	return &task.TaskSortResponse{}, nil
}

// sortTask 对任务进行排序，根据指定的前一个任务和目标阶段重新安排任务的顺序。
func (t *TaskService) sortTask(preTaskCode int64, nextTaskCode string, toStageCode int64) error {
	// 创建一个带有超时的context，以防止排序操作无限期地执行。
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 查找前一个任务，以便获取其排序信息。
	ts, err := t.taskRepo.FindTaskById(c, preTaskCode)
	if err != nil {
		zap.L().Error("project task TaskSort taskRepo.FindTaskById error", zap.Error(err))
		return errs.GrpcError(model.DBError)
	}

	// 使用事务来确保排序操作的原子性。
	err = t.transaction.Action(func(conn database.DbConn) error {
		// 如果任务已经处于正确的位置，则无需进行任何操作。
		ts.StageCode = int(toStageCode)
		if nextTaskCode != "" {
			// 解密nextTaskCode以获取实际的任务代码。
			nextTaskCode := encrypts.DecryptNoErr(nextTaskCode)
			next, err := t.taskRepo.FindTaskById(c, nextTaskCode)
			if err != nil {
				zap.L().Error("project task TaskSort taskRepo.FindTaskById error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			// 查找比next任务排序值小的最大任务，以便插入ts任务。
			prepre, err := t.taskRepo.FindTaskByStageCodeLtSort(c, next.StageCode, next.Sort)
			if err != nil {
				zap.L().Error("project task TaskSort taskRepo.FindTaskByStageCodeLtSort error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			if prepre != nil {
				ts.Sort = (prepre.Sort + next.Sort) / 2
			}
			if prepre == nil {
				ts.Sort = 0
			}
		} else {
			// 如果没有下一个任务代码，将任务排序到阶段的末尾。
			maxSort, err := t.taskRepo.FindTaskSort(c, ts.ProjectCode, int64(ts.StageCode))
			if err != nil {
				zap.L().Error("project task TaskSort taskRepo.FindTaskSort error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			if maxSort == nil {
				a := 0
				maxSort = &a
			}
			ts.Sort = *maxSort + 65536
		}
		// 如果排序值小于50，重置排序值以避免过小的排序值。
		if ts.Sort < 50 {
			err = t.resetSort(toStageCode)
			if err != nil {
				zap.L().Error("project task TaskSort resetSort error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			// 递归调用sortTask以重新尝试排序。
			return t.sortTask(preTaskCode, nextTaskCode, toStageCode)
		}
		// 更新任务的排序值。
		err = t.taskRepo.UpdateTaskSort(c, conn, ts)
		if err != nil {
			zap.L().Error("project task TaskSort taskRepo.UpdateTaskSort error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	})
	return err
}

// resetSort 重置指定阶段的任务排序
// 参数 stageCode: 阶段代码
// 返回 error: 错误信息，如果有的话
func (t *TaskService) resetSort(stageCode int64) error {
	// 获取指定阶段的任务列表
	list, err := t.taskRepo.FindTaskByStageCode(context.Background(), int(stageCode))
	if err != nil {
		return err
	}
	// 通过事务更新任务排序
	return t.transaction.Action(func(conn database.DbConn) error {
		iSort := 65536
		for index, v := range list {
			v.Sort = (index + 1) * iSort
			// 更新数据库中任务的排序值
			return t.taskRepo.UpdateTaskSort(context.Background(), conn, v)
		}
		return nil
	})
}

// MyTaskList 获取我的任务列表
// 参数 ctx: 上下文
// 参数 msg: 任务请求消息，包含任务类型、成员ID、页码和页面大小
// 返回 MyTaskListResponse: 任务列表响应，包含任务列表和总任务数
// 返回 error: 错误信息，如果有的话
func (t *TaskService) MyTaskList(ctx context.Context, msg *task.TaskReqMessage) (*task.MyTaskListResponse, error) {
	var tsList []*data.Task
	var err error
	var total int64
	// 根据任务类型获取任务列表
	if msg.TaskType == 1 {
		// 我执行的任务
		tsList, total, err = t.taskRepo.FindTaskByAssignTo(ctx, msg.MemberId, int(msg.Type), msg.Page, msg.PageSize)
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByAssignTo error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	}
	if msg.TaskType == 2 {
		// 我参与的任务
		tsList, total, err = t.taskRepo.FindTaskByMemberCode(ctx, msg.MemberId, int(msg.Type), msg.Page, msg.PageSize)
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByMemberCode error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	}
	if msg.TaskType == 3 {
		// 我创建的任务
		tsList, total, err = t.taskRepo.FindTaskByCreateBy(ctx, msg.MemberId, int(msg.Type), msg.Page, msg.PageSize)
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByCreateBy error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	}
	// 如果没有找到任务，返回空列表和总任务数0
	if tsList == nil || len(tsList) <= 0 {
		return &task.MyTaskListResponse{List: nil, Total: 0}, nil
	}
	var pids []int64
	var mids []int64
	// 收集任务列表中的项目ID和分配给的成员ID
	for _, v := range tsList {
		pids = append(pids, v.ProjectCode)
		mids = append(mids, v.AssignTo)
	}
	// 获取项目信息
	pList, err := t.projectRepo.FindProjectByIds(ctx, pids)
	projectMap := pro.ToProjectMap(pList)

	// 获取成员信息
	mList, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, &login.UserMessage{
		MIds: mids,
	})
	mMap := make(map[int64]*login.MemberMessage)
	for _, v := range mList.List {
		mMap[v.Id] = v
	}
	var mtdList []*data.MyTaskDisplay
	// 构建我的任务显示列表
	for _, v := range tsList {
		memberMessage := mMap[v.AssignTo]
		name := memberMessage.Name
		avatar := memberMessage.Avatar
		mtd := v.ToMyTaskDisplay(projectMap[v.ProjectCode], name, avatar)
		mtdList = append(mtdList, mtd)
	}
	var myMsgs []*task.MyTaskMessage
	// 复制我的任务显示列表到响应消息
	copier.Copy(&myMsgs, mtdList)
	return &task.MyTaskListResponse{List: myMsgs, Total: total}, nil
}

// ReadTask 读取任务详情
func (t *TaskService) ReadTask(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskMessage, error) {
	// 根据taskCode查询任务详情 根据任务查询项目详情 根据任务查询任务步骤详情 查询任务的执行者的成员详情
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 查询任务详情
	taskInfo, err := t.taskRepo.FindTaskById(c, taskCode)
	if err != nil {
		zap.L().Error("project task ReadTask taskRepo FindTaskById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if taskInfo == nil {
		return &task.TaskMessage{}, nil
	}
	display := taskInfo.ToTaskDisplay()
	if taskInfo.Private == 1 {
		//代表隐私模式
		// FindTaskMemberByTaskId 根据任务ID和成员ID查找任务成员
		taskMember, err := t.taskRepo.FindTaskMemberByTaskId(ctx, taskInfo.Id, msg.MemberId)
		if err != nil {
			zap.L().Error("project task TaskList taskRepo.FindTaskMemberByTaskId error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
		if taskMember != nil {
			display.CanRead = model.CanRead
		} else {
			display.CanRead = model.NoCanRead
		}
	}
	// 查询项目详情
	pj, err := t.projectRepo.FindProjectById(c, taskInfo.ProjectCode)
	display.ProjectName = pj.Name
	// 查询任务步骤详情
	taskStages, err := t.taskStagesRepo.FindById(c, taskInfo.StageCode)
	display.StageName = taskStages.Name
	// 查询任务的执行者的成员详情
	memberMessage, err := rpc.LoginServiceClient.FindMemInfoById(ctx, &login.UserMessage{MemId: taskInfo.AssignTo})
	if err != nil {
		zap.L().Error("project task TaskList LoginServiceClient.FindMemInfoById error", zap.Error(err))
		return nil, err
	}
	// 构建执行者的成员详情
	e := data.Executor{
		Name:   memberMessage.Name,
		Avatar: memberMessage.Avatar,
	}
	display.Executor = e
	var taskMessage = &task.TaskMessage{}
	copier.Copy(taskMessage, display)
	return taskMessage, nil
}

// ListTaskMember 获取任务成员列表
func (t *TaskService) ListTaskMember(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskMemberList, error) {
	//查询 task member表 根据memberCode去查询用户信息
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 查询任务成员列表
	taskMemberPage, total, err := t.taskRepo.FindTaskMemberPage(c, taskCode, msg.Page, msg.PageSize)
	if err != nil {
		zap.L().Error("project task TaskList taskRepo.FindTaskMemberPage error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var mids []int64
	// 遍历任务成员列表，将memberCode添加到mids切片中
	for _, v := range taskMemberPage {
		mids = append(mids, v.MemberCode)
	}
	// 通过成员ids查询成员信息
	messageList, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, &login.UserMessage{MIds: mids})
	mMap := make(map[int64]*login.MemberMessage, len(messageList.List))
	for _, v := range messageList.List {
		mMap[v.Id] = v
	}
	// 遍历任务成员列表，构建TaskMemberMessage对象
	var taskMemeberMemssages []*task.TaskMemberMessage
	for _, v := range taskMemberPage {
		tm := &task.TaskMemberMessage{}
		tm.Code = encrypts.EncryptNoErr(v.MemberCode)
		tm.Id = v.Id
		message := mMap[v.MemberCode]
		tm.Name = message.Name
		tm.Avatar = message.Avatar
		tm.IsExecutor = int32(v.IsExecutor)
		tm.IsOwner = int32(v.IsOwner)
		taskMemeberMemssages = append(taskMemeberMemssages, tm)
	}
	return &task.TaskMemberList{List: taskMemeberMemssages, Total: total}, nil
}

// TaskLog 获取任务日志
func (t *TaskService) TaskLog(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskLogList, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	all := msg.All
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 查询任务日志
	var list []*dataNew.ProjectLog
	var total int64
	var err error
	if all == 1 {
		//显示全部
		// 根据任务编码和评论类型查询任务日志
		list, total, err = t.projectLogRepo.FindLogByTaskCode(c, taskCode, int(msg.Comment))
	}
	if all == 0 {
		//分页
		// 根据任务编码和评论类型分页查询任务日志
		list, total, err = t.projectLogRepo.FindLogByTaskCodePage(c, taskCode, int(msg.Comment), int(msg.Page), int(msg.PageSize))
	}
	if err != nil {
		zap.L().Error("project task TaskLog projectLogRepo.FindLogByTaskCodePage error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if total == 0 {
		return &task.TaskLogList{}, nil
	}
	var displayList []*dataNew.ProjectLogDisplay
	var mIdList []int64
	// 遍历任务日志列表，将memberCode添加到mIdList切片中
	for _, v := range list {
		mIdList = append(mIdList, v.MemberCode)
	}
	// 通过成员ids查询成员信息
	messageList, err := rpc.LoginServiceClient.FindMemInfoByIds(c, &login.UserMessage{MIds: mIdList})
	// 创建成员信息映射
	mMap := make(map[int64]*login.MemberMessage)
	// 遍历成员信息列表，将成员信息添加到mMap中
	for _, v := range messageList.List {
		mMap[v.Id] = v
	}
	// 遍历任务日志列表，构建TaskLogDisplay对象
	for _, v := range list {
		// 获取对应日志信息
		display := v.ToDisplay()
		message := mMap[v.MemberCode]
		m := dataNew.Member{}
		m.Name = message.Name
		m.Id = message.Id
		m.Avatar = message.Avatar
		m.Code = message.Code
		display.Member = m
		displayList = append(displayList, display)
	}
	var l []*task.TaskLog
	copier.Copy(&l, displayList)
	return &task.TaskLogList{List: l, Total: total}, nil
}

func (t *TaskService) TaskWorkTimeList(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskWorkTimeResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var list []*data.TaskWorkTime
	var err error
	list, err = t.taskWorkTimeRepo.FindWorkTimeList(c, taskCode)
	if err != nil {
		zap.L().Error("project task TaskWorkTimeList taskWorkTimeRepo.FindWorkTimeList error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if len(list) == 0 {
		return &task.TaskWorkTimeResponse{}, nil
	}
	var displayList []*data.TaskWorkTimeDisplay
	var mIdList []int64
	for _, v := range list {
		mIdList = append(mIdList, v.MemberCode)
	}
	messageList, err := rpc.LoginServiceClient.FindMemInfoByIds(c, &login.UserMessage{MIds: mIdList})
	mMap := make(map[int64]*login.MemberMessage)
	for _, v := range messageList.List {
		mMap[v.Id] = v
	}
	for _, v := range list {
		display := v.ToDisplay()
		message := mMap[v.MemberCode]
		m := data.Member{}
		m.Name = message.Name
		m.Id = message.Id
		m.Avatar = message.Avatar
		m.Code = message.Code
		display.Member = m
		displayList = append(displayList, display)
	}
	var l []*task.TaskWorkTime
	copier.Copy(&l, displayList)
	return &task.TaskWorkTimeResponse{List: l, Total: int64(len(l))}, nil
}
func (t *TaskService) SaveTaskWorkTime(ctx context.Context, msg *task.TaskReqMessage) (*task.SaveTaskWorkTimeResponse, error) {
	tmt := &data.TaskWorkTime{}
	tmt.BeginTime = msg.BeginTime
	tmt.Num = int(msg.Num)
	tmt.Content = msg.Content
	tmt.TaskCode = encrypts.DecryptNoErr(msg.TaskCode)
	tmt.MemberCode = msg.MemberId
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := t.taskWorkTimeRepo.Save(c, tmt)
	if err != nil {
		zap.L().Error("project task SaveTaskWorkTime taskWorkTimeRepo.Save error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &task.SaveTaskWorkTimeResponse{}, nil
}

func (t *TaskService) SaveTaskFile(ctx context.Context, msg *task.TaskFileReqMessage) (*task.TaskFileResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	//存file表
	f := &data.File{
		PathName:         msg.PathName,
		Title:            msg.FileName,
		Extension:        msg.Extension,
		Size:             int(msg.Size),
		ObjectType:       "",
		OrganizationCode: encrypts.DecryptNoErr(msg.OrganizationCode),
		TaskCode:         encrypts.DecryptNoErr(msg.TaskCode),
		ProjectCode:      encrypts.DecryptNoErr(msg.ProjectCode),
		CreateBy:         msg.MemberId,
		CreateTime:       time.Now().UnixMilli(),
		Downloads:        0,
		Extra:            "",
		Deleted:          model.NoDeleted,
		FileType:         msg.FileType,
		FileUrl:          msg.FileUrl,
		DeletedTime:      0,
	}
	err := t.fileRepo.Save(context.Background(), f)
	if err != nil {
		zap.L().Error("project task SaveTaskFile fileRepo.Save error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	//存入source_link
	sl := &data.SourceLink{
		SourceType:       "file",
		SourceCode:       f.Id,
		LinkType:         "task",
		LinkCode:         taskCode,
		OrganizationCode: encrypts.DecryptNoErr(msg.OrganizationCode),
		CreateBy:         msg.MemberId,
		CreateTime:       time.Now().UnixMilli(),
		Sort:             0,
	}
	err = t.sourceLinkRepo.Save(context.Background(), sl)
	if err != nil {
		zap.L().Error("project task SaveTaskFile sourceLinkRepo.Save error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &task.TaskFileResponse{}, nil
}

func (t *TaskService) TaskSources(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskSourceResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	sourceLinks, err := t.sourceLinkRepo.FindByTaskCode(context.Background(), taskCode)
	if err != nil {
		zap.L().Error("project task SaveTaskFile sourceLinkRepo.FindByTaskCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if len(sourceLinks) == 0 {
		return &task.TaskSourceResponse{}, nil
	}
	var fIdList []int64
	for _, v := range sourceLinks {
		fIdList = append(fIdList, v.SourceCode)
	}
	files, err := t.fileRepo.FindByIds(context.Background(), fIdList)
	if err != nil {
		zap.L().Error("project task SaveTaskFile fileRepo.FindByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	fMap := make(map[int64]*data.File)
	for _, v := range files {
		fMap[v.Id] = v
	}
	var list []*data.SourceLinkDisplay
	for _, v := range sourceLinks {
		list = append(list, v.ToDisplay(fMap[v.SourceCode]))
	}
	var slMsg []*task.TaskSourceMessage
	copier.Copy(&slMsg, list)
	return &task.TaskSourceResponse{List: slMsg}, nil
}

func (t *TaskService) CreateComment(ctx context.Context, msg *task.TaskReqMessage) (*task.CreateCommentResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	taskById, err := t.taskRepo.FindTaskById(context.Background(), taskCode)
	if err != nil {
		zap.L().Error("project task CreateComment fileRepo.FindTaskById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	pl := &data.ProjectLog{
		MemberCode:   msg.MemberId,
		Content:      msg.CommentContent,
		Remark:       msg.CommentContent,
		Type:         "createComment",
		CreateTime:   time.Now().UnixMilli(),
		SourceCode:   taskCode,
		ActionType:   "task",
		ToMemberCode: 0,
		IsComment:    model.Comment,
		ProjectCode:  taskById.ProjectCode,
		Icon:         "plus",
		IsRobot:      0,
	}
	t.projectLogRepo.SaveProjectLog(pl)
	return &task.CreateCommentResponse{}, nil
}
