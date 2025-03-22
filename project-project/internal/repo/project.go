package repo

import (
	"context"
	"project-project/internal/data/pro"
	"project-project/internal/database"
)

type ProjectRepo interface {
	// FindProjectByMemId 根据成员id查询项目
	FindProjectByMemId(ctx context.Context, memId int64, condition string, page int64, size int64) ([]*pro.ProjectAndMember, int64, error)
	// FindCollectProjectByMemId 查询收藏项目
	FindCollectProjectByMemId(ctx context.Context, memberId int64, page int64, size int64) ([]*pro.ProjectAndMember, int64, error)
	// SaveProject 保存项目
	SaveProject(conn database.DbConn, ctx context.Context, pr *pro.Project) error
	// SaveProjectMember 保存项目成员
	SaveProjectMember(conn database.DbConn, ctx context.Context, pm *pro.ProjectMember) error
	// FindProjectByPIdAndMemId 根据项目id和成员id查询项目
	FindProjectByPIdAndMemId(ctx context.Context, projectCode int64, memberId int64) (*pro.ProjectAndMember, error)
	// FindCollectByPidAndMemId 根据项目id和成员id查询项目收藏
	FindCollectByPidAndMemId(ctx context.Context, projectCode int64, memberId int64) (bool, error)
	// UpdateDeletedProject 更新项目删除状态
	UpdateDeletedProject(ctx context.Context, code int64, deleted bool) error
	// SaveProjectCollect 保存项目收藏
	SaveProjectCollect(ctx context.Context, pc *pro.ProjectCollection) error
	// DeleteProjectCollect 删除项目收藏
	DeleteProjectCollect(ctx context.Context, memId int64, projectCode int64) error
	// UpdateProject 更新项目
	UpdateProject(ctx context.Context, proj *pro.Project) error
	// FindProjectMemberByPid 查询项目成员通过项目code
	FindProjectMemberByPid(ctx context.Context, projectCode int64) (list []*pro.ProjectMember, total int64, err error)
	// FindProjectById 查询项目通过项目id
	FindProjectById(ctx context.Context, projectCode int64) (pj *pro.Project, err error)
	// FindProjectByIds 查询项目通过项目id集合
	FindProjectByIds(ctx context.Context, pids []int64) (list []*pro.Project, err error)
}

type ProjectTemplateRepo interface {
	// 查询系统模板
	FindProjectTemplateSystem(ctx context.Context, page int64, size int64) ([]pro.ProjectTemplate, int64, error)
	// 查询自定义模板
	FindProjectTemplateCustom(ctx context.Context, memId int64, organizationCode int64, page int64, size int64) ([]pro.ProjectTemplate, int64, error)
	// 查询所有模板
	FindProjectTemplateAll(ctx context.Context, organizationCode int64, page int64, size int64) ([]pro.ProjectTemplate, int64, error)
}
