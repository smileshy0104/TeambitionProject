package project_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"project-common/encrypts"
	"project-common/errs"
	"project-common/tms"
	"project-grpc/project"
	"project-grpc/user/login"
	"project-project/internal/dao"
	"project-project/internal/data"
	"project-project/internal/data/menu"
	"project-project/internal/data/pro"
	"project-project/internal/data/task"
	"project-project/internal/database"
	"project-project/internal/database/tran"
	"project-project/internal/domain"
	"project-project/internal/repo"
	"project-project/internal/rpc"
	"project-project/pkg/model"
	"strconv"
	"time"
)

// ProjectService 项目服务结构体，实现了项目服务的服务器接口
// 它通过组合多个仓库和缓存来提供项目相关的服务
type ProjectService struct {
	project.UnimplementedProjectServiceServer                             // 嵌入未实现的项目服务服务器接口
	cache                                     repo.Cache                  // 缓存接口，用于快速数据访问
	transaction                               tran.Transaction            // 事务接口，用于处理数据库事务
	menuRepo                                  repo.MenuRepo               // 菜单仓库接口，用于菜单相关操作
	projectRepo                               repo.ProjectRepo            // 项目仓库接口，用于项目相关操作
	projectTemplateRepo                       repo.ProjectTemplateRepo    // 项目模板仓库接口，用于项目模板相关操作
	taskStagesTemplateRepo                    repo.TaskStagesTemplateRepo // 任务阶段模板仓库接口，用于任务阶段模板相关操作
	taskStagesRepo                            repo.TaskStagesRepo
	projectLogRepo                            repo.ProjectLogRepo
	taskRepo                                  repo.TaskRepo
	nodeDomain                                *domain.ProjectNodeDomain
	taskDomain                                *domain.TaskDomain
}

// New 创建并返回一个新的 ProjectService 实例
// 它初始化了 ProjectService 结构体，并注入了必要的依赖
func New() *ProjectService {
	return &ProjectService{
		cache:                  dao.Rc,                         // 使用全局缓存实例
		transaction:            dao.NewTransaction(),           // 创建新的事务实例
		menuRepo:               dao.NewMenuDao(),               // 创建新的菜单仓库实例
		projectRepo:            dao.NewProjectDao(),            // 创建新的项目仓库实例
		projectTemplateRepo:    dao.NewProjectTemplateDao(),    // 创建新的项目模板仓库实例
		taskStagesTemplateRepo: dao.NewTaskStagesTemplateDao(), // 创建新的任务阶段模板仓库实例
		taskStagesRepo:         dao.NewTaskStagesDao(),
		projectLogRepo:         dao.NewProjectLogDao(),
		taskRepo:               dao.NewTaskDao(),
		nodeDomain:             domain.NewProjectNodeDomain(),
		taskDomain:             domain.NewTaskDomain(),
	}
}

// Index 获取项目的菜单列表
// 该方法从数据库中检索菜单信息，并将其转换为菜单消息列表返回
func (p *ProjectService) Index(context.Context, *project.IndexMessage) (*project.IndexResponse, error) {
	// 从数据库中获取菜单列表
	pms, err := p.menuRepo.FindMenus(context.Background())
	if err != nil {
		// 如果发生错误，记录错误日志并返回数据库错误
		zap.L().Error("Index db FindMenus error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	// 将获取的菜单信息转换为子菜单结构
	childs := menu.CovertChild(pms)
	// 初始化菜单消息列表
	var mms []*project.MenuMessage
	// 将子菜单结构复制到菜单消息列表中
	copier.Copy(&mms, childs)
	// 返回包含菜单列表的响应对象
	return &project.IndexResponse{Menus: mms}, nil
}

// FindProjectByMemId 根据成员ID查找项目信息。
// 该方法根据成员ID和查询条件（如"my", "archive", "deleted", "collect"）来获取相应的项目列表和总数。
// 它还处理了项目加密、访问控制类型和时间格式化等逻辑。
func (p *ProjectService) FindProjectByMemId(ctx context.Context, msg *project.ProjectRpcMessage) (*project.MyProjectResponse, error) {
	memberId := msg.MemberId
	page := msg.Page
	pageSize := msg.PageSize
	var pms []*pro.ProjectAndMember
	var total int64
	var err error
	// 根据选择的类型查询项目
	if msg.SelectBy == "" || msg.SelectBy == "my" {
		pms, total, err = p.projectRepo.FindProjectByMemId(ctx, memberId, "and deleted=0 ", page, pageSize)
	}
	// 查询归档的项目
	if msg.SelectBy == "archive" {
		pms, total, err = p.projectRepo.FindProjectByMemId(ctx, memberId, "and archive=1 ", page, pageSize)
	}
	// 查询回收站中的项目
	if msg.SelectBy == "deleted" {
		pms, total, err = p.projectRepo.FindProjectByMemId(ctx, memberId, "and deleted=1 ", page, pageSize)
	}
	// 查询收藏的项目
	if msg.SelectBy == "collect" {
		pms, total, err = p.projectRepo.FindCollectProjectByMemId(ctx, memberId, page, pageSize)
		for _, v := range pms {
			v.Collected = model.Collected
		}
	} else {
		collectPms, _, err := p.projectRepo.FindCollectProjectByMemId(ctx, memberId, page, pageSize)
		if err != nil {
			zap.L().Error("project FindProjectByMemId::FindCollectProjectByMemId error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
		var cMap = make(map[int64]*pro.ProjectAndMember)
		for _, v := range collectPms {
			cMap[v.Id] = v
		}
		for _, v := range pms {
			if cMap[v.ProjectCode] != nil {
				v.Collected = model.Collected
			}
		}
	}
	if err != nil {
		zap.L().Error("project FindProjectByMemId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if pms == nil {
		return &project.MyProjectResponse{Pm: []*project.ProjectMessage{}, Total: total}, nil
	}

	var pmm []*project.ProjectMessage
	copier.Copy(&pmm, pms)
	for _, v := range pmm {
		v.Code, _ = encrypts.EncryptInt64(v.ProjectCode, model.AESKey)
		pam := pro.ToMap(pms)[v.Id]
		v.AccessControlType = pam.GetAccessControlType()
		v.OrganizationCode, _ = encrypts.EncryptInt64(pam.OrganizationCode, model.AESKey)
		v.JoinTime = tms.FormatByMill(pam.JoinTime)
		v.OwnerName = msg.MemberName
		v.Order = int32(pam.Sort)
		v.CreateTime = tms.FormatByMill(pam.CreateTime)
	}
	return &project.MyProjectResponse{Pm: pmm, Total: total}, nil
}

// FindProjectTemplate 根据不同的查看类型查找项目模板。
// 该方法首先根据加密的组织代码查询相关的项目模板，然后根据不同的查看类型
// 查询所有模板、自定义模板或系统模板。查询到的模板信息随后用于获取相应的任务阶段模板，
// 最终将这些信息组装成项目模板响应返回。
func (ps *ProjectService) FindProjectTemplate(ctx context.Context, msg *project.ProjectRpcMessage) (*project.ProjectTemplateResponse, error) {
	// 解密组织代码并转换为int64类型
	organizationCodeStr, _ := encrypts.Decrypt(msg.OrganizationCode, model.AESKey)
	organizationCode, _ := strconv.ParseInt(organizationCodeStr, 10, 64)
	page := msg.Page
	pageSize := msg.PageSize
	// 定义项目模板数组和总数变量
	var pts []pro.ProjectTemplate
	var total int64
	var err error

	// 根据不同的查看类型查询项目模板
	if msg.ViewType == -1 {
		pts, total, err = ps.projectTemplateRepo.FindProjectTemplateAll(ctx, organizationCode, page, pageSize)
	}
	if msg.ViewType == 0 {
		pts, total, err = ps.projectTemplateRepo.FindProjectTemplateCustom(ctx, msg.MemberId, organizationCode, page, pageSize)
	}
	if msg.ViewType == 1 {
		pts, total, err = ps.projectTemplateRepo.FindProjectTemplateSystem(ctx, page, pageSize)
	}
	if err != nil {
		zap.L().Error("project FindProjectTemplate FindProjectTemplateSystem error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}

	// 根据项目模板ID查询任务阶段模板
	tsts, err := ps.taskStagesTemplateRepo.FindInProTemIds(ctx, pro.ToProjectTemplateIds(pts))
	if err != nil {
		zap.L().Error("project FindProjectTemplate FindInProTemIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}

	// 将查询到的项目模板和任务阶段模板进行组装
	var ptas []*pro.ProjectTemplateAll
	for _, v := range pts {
		// 将项目模板和任务阶段模板进行组装
		ptas = append(ptas, v.Convert(task.CovertProjectMap(tsts)[v.Id]))
	}

	// 将组装好的数据转换为目标结构并返回
	var pmMsgs []*project.ProjectTemplateMessage
	copier.Copy(&pmMsgs, ptas)
	return &project.ProjectTemplateResponse{Ptm: pmMsgs, Total: total}, nil
}

// SaveProject 保存项目信息及其成员关联
// 该方法首先解密组织代码和模板代码，然后保存项目信息和项目成员关联信息到数据库
func (ps *ProjectService) SaveProject(ctx context.Context, msg *project.ProjectRpcMessage) (*project.SaveProjectMessage, error) {
	// 解密组织代码
	organizationCodeStr, _ := encrypts.Decrypt(msg.OrganizationCode, model.AESKey)
	// 将解密后的组织代码转换为int64类型
	organizationCode, _ := strconv.ParseInt(organizationCodeStr, 10, 64)
	// 解密模板代码
	templateCodeStr, _ := encrypts.Decrypt(msg.TemplateCode, model.AESKey)
	// 将解密后的模板代码转换为int64类型
	templateCode, _ := strconv.ParseInt(templateCodeStr, 10, 64)
	//1. 保存项目表
	pr := &pro.Project{
		Name:              msg.Name,
		Description:       msg.Description,
		TemplateCode:      int(templateCode),
		CreateTime:        time.Now().UnixMilli(),
		Cover:             "https://img2.baidu.com/it/u=792555388,2449797505&fm=253&fmt=auto&app=138&f=JPEG?w=667&h=500",
		Deleted:           model.NoDeleted,
		Archive:           model.NoArchive,
		OrganizationCode:  organizationCode,
		AccessControlType: model.Open,
		TaskBoardTheme:    model.Simple,
	}
	// 使用事务保存项目和项目成员关联
	err := ps.transaction.Action(func(conn database.DbConn) error {
		// 保存项目信息
		err := ps.projectRepo.SaveProject(conn, ctx, pr)
		if err != nil {
			zap.L().Error("project SaveProject SaveProject error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		// 创建项目成员关联对象
		pm := &pro.ProjectMember{
			ProjectCode: pr.Id,
			MemberCode:  msg.MemberId,
			JoinTime:    time.Now().UnixMilli(),
			IsOwner:     msg.MemberId,
			Authorize:   "",
		}
		//2. 保存项目和成员的关联表
		err = ps.projectRepo.SaveProjectMember(conn, ctx, pm)
		if err != nil {
			zap.L().Error("project SaveProject SaveProjectMember error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// 加密项目代码
	code, _ := encrypts.EncryptInt64(pr.Id, model.AESKey)
	// 构建返回消息
	rsp := &project.SaveProjectMessage{
		Id:               pr.Id,
		Code:             code,
		OrganizationCode: organizationCodeStr,
		Name:             pr.Name,
		Cover:            pr.Cover,
		CreateTime:       tms.FormatByMill(pr.CreateTime),
		TaskBoardTheme:   pr.TaskBoardTheme,
	}
	return rsp, nil
}

// 1. 查项目表
// 2. 项目和成员的关联表 查到项目的拥有者 去member表查名字
// 3. 查收藏表 判断收藏状态
// FindProjectDetail 根据项目代码和成员ID查找项目详情。
// 该方法首先解密项目代码，并解析为整数类型。
// 然后，它通过项目ID和成员ID查询项目和成员的详细信息。
// 如果查询成功，它会通过RPC调用获取项目所有者的详细信息，并检查当前成员是否收藏了该项目。
// 最后，它将查询到的信息组装成项目详情消息并返回。
func (ps *ProjectService) FindProjectDetail(ctx context.Context, msg *project.ProjectRpcMessage) (*project.ProjectDetailMessage, error) {
	// 解密项目代码并解析为整数类型。
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	memberId := msg.MemberId

	// 创建一个带有超时的上下文，以防止长时间运行的查询导致系统资源耗尽。
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 查询项目和成员的详细信息。
	projectAndMember, err := ps.projectRepo.FindProjectByPIdAndMemId(c, projectCode, memberId)
	if err != nil {
		zap.L().Error("project FindProjectDetail FindProjectByPIdAndMemId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}

	// 获取项目所有者的ID，并通过RPC调用查询所有者的详细信息。
	ownerId := projectAndMember.IsOwner
	member, err := rpc.LoginServiceClient.FindMemInfoById(c, &login.UserMessage{MemId: ownerId})
	if err != nil {
		zap.L().Error("project rpc FindProjectDetail FindMemInfoById error", zap.Error(err))
		return nil, err
	}

	// 检查当前成员是否收藏了该项目。
	isCollect, err := ps.projectRepo.FindCollectByPidAndMemId(c, projectCode, memberId)
	if err != nil {
		zap.L().Error("project FindProjectDetail FindCollectByPidAndMemId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if isCollect {
		projectAndMember.Collected = model.Collected
	}

	// 初始化项目详情消息，并将查询到的信息复制到该消息中。
	var detailMsg = &project.ProjectDetailMessage{}
	copier.Copy(detailMsg, projectAndMember)

	// 设置项目所有者的头像和名称。
	detailMsg.OwnerAvatar = member.Avatar
	detailMsg.OwnerName = member.Name

	// 加密项目ID和组织代码，以保护敏感信息。
	detailMsg.Code, _ = encrypts.EncryptInt64(projectAndMember.Id, model.AESKey)
	detailMsg.OrganizationCode, _ = encrypts.EncryptInt64(projectAndMember.OrganizationCode, model.AESKey)

	// 设置项目的排序顺序和创建时间。
	detailMsg.Order = int32(projectAndMember.Sort)
	detailMsg.CreateTime = tms.FormatByMill(projectAndMember.CreateTime)

	// 返回项目详情消息。
	return detailMsg, nil
}

// UpdateDeletedProject 更新项目删除状态。
// 该函数接收一个ProjectRpcMessage对象，从中解密项目代码并将其转换为整数类型。
// 然后，它使用项目代码和删除状态更新数据库中的项目信息。
// 如果更新成功，返回一个空的DeletedProjectResponse对象；如果更新失败，返回一个错误。
func (ps *ProjectService) UpdateDeletedProject(ctx context.Context, msg *project.ProjectRpcMessage) (*project.DeletedProjectResponse, error) {
	// 解密项目代码。
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	// 将解密后的项目代码字符串转换为整数类型。
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	// 创建一个带有超时的context，以确保数据库操作不会无限期地等待。
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 调用项目仓库的UpdateDeletedProject方法更新数据库中的项目删除状态。
	err := ps.projectRepo.UpdateDeletedProject(c, projectCode, msg.Deleted)
	if err != nil {
		// 如果更新失败，记录错误日志并返回一个gRPC错误。
		zap.L().Error("project RecycleProject DeleteProject error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	// 如果更新成功，返回一个空的DeletedProjectResponse对象。
	return &project.DeletedProjectResponse{}, nil
}

// UpdateProject 更新项目信息。
// 该方法接收一个更新项目的请求消息，解密项目代码并解析为整数类型，
// 然后使用解析后的项目代码和其他消息信息来更新数据库中的项目记录。
// 如果更新过程中发生错误，会记录错误日志并返回一个gRPC错误。
func (ps *ProjectService) UpdateProject(ctx context.Context, msg *project.UpdateProjectMessage) (*project.UpdateProjectResponse, error) {
	// 解密项目代码
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	// 将解密后的项目代码解析为int64类型
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	// 创建一个带有超时的context，以防止更新操作长时间运行
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 创建一个Project实例，填充从消息中获取的更新信息
	proj := &pro.Project{
		Id:                 projectCode,
		Name:               msg.Name,
		Description:        msg.Description,
		Cover:              msg.Cover,
		TaskBoardTheme:     msg.TaskBoardTheme,
		Prefix:             msg.Prefix,
		Private:            int(msg.Private),
		OpenPrefix:         int(msg.OpenPrefix),
		OpenBeginTime:      int(msg.OpenBeginTime),
		OpenTaskPrivate:    int(msg.OpenTaskPrivate),
		Schedule:           msg.Schedule,
		AutoUpdateSchedule: int(msg.AutoUpdateSchedule),
	}
	// 调用项目仓库的更新方法来更新项目信息
	err := ps.projectRepo.UpdateProject(c, proj)
	if err != nil {
		// 如果更新过程中发生错误，记录错误日志并返回一个gRPC错误
		zap.L().Error("project UpdateProject::UpdateProject error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	// 返回一个空的更新项目响应，表示更新操作成功
	return &project.UpdateProjectResponse{}, nil
}

func (ps *ProjectService) GetLogBySelfProject(ctx context.Context, msg *project.ProjectRpcMessage) (*project.ProjectLogResponse, error) {
	//根据用户id查询当前的用户的日志表

	projectLogs, total, err := ps.projectLogRepo.FindLogByMemberCode(context.Background(), msg.MemberId, msg.Page, msg.PageSize)
	if err != nil {
		zap.L().Error("project ProjectService::GetLogBySelfProject projectLogRepo.FindLogByMemberCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	//查询项目信息
	pIdList := make([]int64, len(projectLogs))
	mIdList := make([]int64, len(projectLogs))
	taskIdList := make([]int64, len(projectLogs))
	for _, v := range projectLogs {
		pIdList = append(pIdList, v.ProjectCode)
		mIdList = append(mIdList, v.MemberCode)
		taskIdList = append(taskIdList, v.SourceCode)
	}
	projects, err := ps.projectRepo.FindProjectByIds(context.Background(), pIdList)
	if err != nil {
		zap.L().Error("project ProjectService::GetLogBySelfProject projectLogRepo.FindProjectByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	pMap := make(map[int64]*data.Project)
	for _, v := range projects {
		pMap[v.Id] = v
	}
	messageList, _ := rpc.LoginServiceClient.FindMemInfoByIds(context.Background(), &login.UserMessage{MIds: mIdList})
	mMap := make(map[int64]*login.MemberMessage)
	for _, v := range messageList.List {
		mMap[v.Id] = v
	}
	tasks, err := ps.taskRepo.FindTaskByIds(context.Background(), taskIdList)
	if err != nil {
		zap.L().Error("project ProjectService::GetLogBySelfProject projectLogRepo.FindTaskByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	tMap := make(map[int64]*data.Task)
	for _, v := range tasks {
		tMap[v.Id] = v
	}
	var list []*data.IndexProjectLogDisplay
	for _, v := range projectLogs {
		display := v.ToIndexDisplay()
		display.ProjectName = pMap[v.ProjectCode].Name
		display.MemberAvatar = mMap[v.MemberCode].Avatar
		display.MemberName = mMap[v.MemberCode].Name
		display.TaskName = tMap[v.SourceCode].Name
		list = append(list, display)
	}
	var msgList []*project.ProjectLogMessage
	copier.Copy(&msgList, list)
	return &project.ProjectLogResponse{List: msgList, Total: total}, nil
}

func (ps *ProjectService) FindProjectByMemberId(ctx context.Context, msg *project.ProjectRpcMessage) (*project.FindProjectByMemberIdResponse, error) {
	isProjectCode := false
	var projectId int64
	if msg.ProjectCode != "" {
		projectId = encrypts.DecryptNoErr(msg.ProjectCode)
		isProjectCode = true
	}
	isTaskCode := false
	var taskId int64
	if msg.TaskCode != "" {
		taskId = encrypts.DecryptNoErr(msg.TaskCode)
		isTaskCode = true
	}
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if !isProjectCode && isTaskCode {
		projectCode, ok, bError := ps.taskDomain.FindProjectIdByTaskId(taskId)
		if bError != nil {
			return nil, bError
		}
		if !ok {
			return &project.FindProjectByMemberIdResponse{
				Project:  nil,
				IsOwner:  false,
				IsMember: false,
			}, nil
		}
		projectId = projectCode
		isProjectCode = true
	}
	if isProjectCode {
		//根据projectid和memberid查询
		pm, err := ps.projectRepo.FindProjectByPIdAndMemId(c, projectId, msg.MemberId)
		if err != nil {
			return nil, model.DBError
		}
		if pm == nil {
			return &project.FindProjectByMemberIdResponse{
				Project:  nil,
				IsOwner:  false,
				IsMember: false,
			}, nil
		}
		projectMessage := &project.ProjectMessage{}
		copier.Copy(projectMessage, pm)
		isOwner := false
		if pm.IsOwner == 1 {
			isOwner = true
		}
		return &project.FindProjectByMemberIdResponse{
			Project:  projectMessage,
			IsOwner:  isOwner,
			IsMember: true,
		}, nil
	}
	return &project.FindProjectByMemberIdResponse{}, nil
}
