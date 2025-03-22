package project_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"project-common/encrypts"
	"project-common/errs"
	"project-grpc/project"
	"project-project/internal/dao"
	"project-project/internal/data/menu"
	"project-project/internal/database/tran"
	"project-project/internal/repo"
	"project-project/pkg/model"
)

// ProjectService 提供了项目相关的服务功能，继承了 project.UnimplementedProjectServiceServer
// 它通过组合多个仓库（缓存、事务、菜单、项目）来实现其功能
type ProjectService struct {
	project.UnimplementedProjectServiceServer                  // 继承自项目服务的未实现服务器，为实现接口做准备
	cache                                     repo.Cache       // 缓存库，用于快速访问频繁请求的数据
	transaction                               tran.Transaction // 事务管理，确保数据库操作的一致性和完整性
	menuRepo                                  repo.MenuRepo    // 菜单仓库，处理与菜单相关的数据访问
	projectRepo                               repo.ProjectRepo // 项目仓库，处理与项目相关的数据访问
}

// New 创建并返回一个新的 ProjectService 实例
// 它初始化了 ProjectService 结构体，并注入了必要的依赖
func New() *ProjectService {
	// 使用 dao（数据访问对象）提供的方法获取必要的实例
	// 这些实例包括缓存、事务管理以及菜单和项目的数据访问对象
	return &ProjectService{
		cache:       dao.Rc,               // 使用全局缓存实例
		transaction: dao.NewTransaction(), // 创建新的事务管理实例
		menuRepo:    dao.NewMenuDao(),     // 创建新的菜单数据访问实例
		projectRepo: dao.NewProjectDao(),  // 创建新的项目数据访问实例
	}
}

// Index 获取项目的菜单列表
// 该方法从数据库中检索菜单信息，并将其转换为菜单消息列表返回
// 参数:
//   - ctx context.Context: 上下文对象，用于传递请求范围的 deadline、取消信号等
//   - msg *project.IndexMessage: 请求消息，可能包含检索菜单所需的参数（当前未使用）
//
// 返回值:
//   - *project.IndexResponse: 包含菜单列表的响应对象
//   - error: 错误对象，如果执行过程中遇到错误则返回
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
// 该方法接收一个上下文和一个包含成员ID、页码和页面大小的RPC消息，
// 并返回一个包含项目信息列表和总项目数的响应对象。
// 主要功能包括：
// - 从RPC消息中提取成员ID、页码和页面大小。
// - 调用项目仓库的FindProjectByMemId方法查询项目信息。
// - 如果查询出错，记录错误日志并返回gRPC错误。
// - 如果查询结果为空，返回一个空项目列表和总项目数的响应对象。
// - 否则，复制查询结果到一个新的项目消息列表，并对每个项目的ID进行加密。
// - 最后，返回包含加密后项目信息列表和总项目数的响应对象。
func (p *ProjectService) FindProjectByMemId(ctx context.Context, msg *project.ProjectRpcMessage) (*project.MyProjectResponse, error) {
	// 提取RPC消息中的成员ID、页码和页面大小。
	memberId := msg.MemberId
	page := msg.Page
	pageSize := msg.PageSize

	// 调用项目仓库的FindProjectByMemId方法查询项目信息。
	pms, total, err := p.projectRepo.FindProjectByMemId(ctx, memberId, page, pageSize)
	if err != nil {
		// 如果查询出错，记录错误日志并返回gRPC错误。
		zap.L().Error("project FindProjectByMemId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}

	// 如果查询结果为空，返回一个空项目列表和总项目数的响应对象。
	if pms == nil {
		return &project.MyProjectResponse{Pm: []*project.ProjectMessage{}, Total: total}, nil
	}

	// 复制查询结果到一个新的项目消息列表。
	var pmm []*project.ProjectMessage
	copier.Copy(&pmm, pms)

	// 对每个项目的ID进行加密。
	for _, v := range pmm {
		v.Code, _ = encrypts.EncryptInt64(v.Id, model.AESKey)
	}

	// 返回包含加密后项目信息列表和总项目数的响应对象。
	return &project.MyProjectResponse{Pm: pmm, Total: total}, nil
}
