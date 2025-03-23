package repo

import (
	"context"
	"project-project/internal/data"
)

type ProjectAuthRepo interface {
	// FindAuthList 查询项目权限列表
	FindAuthList(ctx context.Context, orgCode int64) (list []*data.ProjectAuth, err error)
	// FindAuthListPage 查询项目权限列表分页
	FindAuthListPage(ctx context.Context, orgCode int64, page int64, pageSize int64) (list []*data.ProjectAuth, total int64, err error)
}
