package repo

import (
	"context"
	"project-project/internal/data/task"
	"project-project/internal/database"
)

// TaskStagesTemplateRepo 是一个接口，定义了与任务阶段模板相关的数据访问方法。
type TaskStagesTemplateRepo interface {
	// FindInProTemIds 根据提供的项目模板ID列表，查找对应的任务阶段模板。
	// 该方法接收一个上下文和一个整数ID列表，返回一个任务阶段模板的列表和一个错误对象。
	FindInProTemIds(ctx context.Context, ids []int) ([]task.MsTaskStagesTemplate, error)

	// FindByProjectTemplateId 根据项目模板ID查找所有的任务阶段模板。
	// 该方法接收一个上下文和一个项目模板代码（整数），返回一个任务阶段模板的指针列表和一个错误对象。
	FindByProjectTemplateId(ctx context.Context, projectTemplateCode int) (list []*task.MsTaskStagesTemplate, err error)
}

// TaskStagesRepo 是一个接口，定义了操作任务阶段数据的通用方法。
type TaskStagesRepo interface {
	// SaveTaskStages 保存任务阶段信息到数据库中。
	// 该方法需要上下文ctx，数据库连接conn，以及待保存的任务阶段对象ts。
	// 返回可能发生的错误。
	SaveTaskStages(ctx context.Context, conn database.DbConn, ts *task.TaskStages) error

	// FindStagesByProjectId 根据项目ID查找对应的任务阶段列表。
	// 该方法需要上下文ctx，项目代码projectCode，当前页码page和每页大小pageSize。
	// 返回一个任务阶段列表list，总记录数total，以及可能发生的错误err。
	FindStagesByProjectId(ctx context.Context, projectCode int64, page int64, pageSize int64) (list []*task.TaskStages, total int64, err error)

	// FindById 根据任务阶段ID查找任务阶段信息。
	// 该方法需要上下文ctx和任务阶段ID id。
	// 返回找到的任务阶段对象ts，以及可能发生的错误err。
	FindById(ctx context.Context, id int) (ts *task.TaskStages, err error)
}

// TaskRepo 定义了任务相关的数据访问接口
type TaskRepo interface {
	// FindTaskByStageCode 根据阶段代码查找任务
	FindTaskByStageCode(ctx context.Context, stageCode int) (list []*task.Task, err error)

	// FindTaskMemberByTaskId 根据任务ID和成员ID查找任务成员
	FindTaskMemberByTaskId(ctx context.Context, taskCode int64, memberId int64) (task *task.TaskMember, err error)

	// FindTaskMaxIdNum 查找项目下任务的最大ID号码
	FindTaskMaxIdNum(ctx context.Context, projectCode int64) (v *int, err error)

	// FindTaskSort 查找任务的排序号
	FindTaskSort(ctx context.Context, projectCode int64, stageCode int64) (v *int, err error)

	// SaveTask 保存任务
	SaveTask(ctx context.Context, conn database.DbConn, ts *task.Task) error

	// SaveTaskMember 保存任务成员
	SaveTaskMember(ctx context.Context, conn database.DbConn, tm *task.TaskMember) error

	// FindTaskById 根据任务代码查找任务
	FindTaskById(ctx context.Context, taskCode int64) (ts *task.Task, err error)

	// UpdateTaskSort 更新任务的排序号
	UpdateTaskSort(ctx context.Context, conn database.DbConn, ts *task.Task) error

	// FindTaskByStageCodeLtSort 查找同一阶段内排序号较小的任务
	FindTaskByStageCodeLtSort(ctx context.Context, stageCode int, sort int) (ts *task.Task, err error)

	// FindTaskByAssignTo 根据分配给的成员ID查找任务
	FindTaskByAssignTo(ctx context.Context, memberId int64, done int, page int64, size int64) ([]*task.Task, int64, error)

	// FindTaskByMemberCode 根据成员代码查找任务
	FindTaskByMemberCode(ctx context.Context, memberId int64, done int, page int64, size int64) (tList []*task.Task, total int64, err error)

	// FindTaskByCreateBy 根据创建者的成员代码查找任务
	FindTaskByCreateBy(ctx context.Context, memberId int64, done int, page int64, size int64) (tList []*task.Task, total int64, err error)
}
