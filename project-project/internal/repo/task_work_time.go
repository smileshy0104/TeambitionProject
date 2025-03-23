package repo

import (
	"context"
	"project-project/internal/data"
)

type TaskWorkTimeRepo interface {
	// Save 保存任务工时
	Save(ctx context.Context, twt *data.TaskWorkTime) error
	// FindWorkTimeList 查询任务工时列表
	FindWorkTimeList(ctx context.Context, taskCode int64) (list []*data.TaskWorkTime, err error)
}
