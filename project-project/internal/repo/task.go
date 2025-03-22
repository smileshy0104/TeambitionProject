package repo

import (
	"context"
	"project-project/internal/data/task"
)

type TaskStagesTemplateRepo interface {
	// FindInProTemIds 根据项目模板id查询任务阶段
	FindInProTemIds(ctx context.Context, ids []int) ([]task.MsTaskStagesTemplate, error)
}
