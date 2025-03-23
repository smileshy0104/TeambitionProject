package repo

import (
	"context"
	"project-project/internal/data"
)

type ProjectLogRepo interface {
	// FindLogByTaskCode 根据任务ID查询日志（全部）
	FindLogByTaskCode(ctx context.Context, taskCode int64, comment int) (list []*data.ProjectLog, total int64, err error)
	// FindLogByTaskCodePage 根据任务ID查询日志（分页）
	FindLogByTaskCodePage(ctx context.Context, taskCode int64, comment int, page int, pageSize int) (list []*data.ProjectLog, total int64, err error)
	// SaveProjectLog 保存日志
	SaveProjectLog(pl *data.ProjectLog)
	// FindLogByMemberCode 根据成员ID查询日志（分页）
	FindLogByMemberCode(background context.Context, memberId int64, page int64, size int64) (list []*data.ProjectLog, total int64, err error)
}
