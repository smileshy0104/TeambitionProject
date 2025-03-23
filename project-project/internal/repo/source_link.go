package repo

import (
	"context"
	"project-project/internal/data"
)

type SourceLinkRepo interface {
	// Save 关联文件数据
	Save(ctx context.Context, link *data.SourceLink) error
	// FindByTaskCode 根据任务ID查询关联文件
	FindByTaskCode(ctx context.Context, taskCode int64) (list []*data.SourceLink, err error)
}
