package domain

import (
	"context"
	"fmt"
	"project-common/errs"
	"project-common/kafka"
	"project-project/config"
	"project-project/internal/dao"
	"project-project/internal/repo"
	"project-project/pkg/model"
)

type TaskDomain struct {
	taskRepo repo.TaskRepo
}

func NewTaskDomain() *TaskDomain {
	return &TaskDomain{
		taskRepo: dao.NewTaskDao(),
	}
}

func (d *TaskDomain) FindProjectIdByTaskId(taskId int64) (int64, bool, *errs.BError) {
	fmt.Println("FindProjectIdByTaskId")
	config.SendLog(kafka.Info("Find", "TaskDomain.FindProjectIdByTaskId", kafka.FieldMap{
		"taskId": taskId,
	}))
	task, err := d.taskRepo.FindTaskById(context.Background(), taskId)
	if err != nil {
		config.SendLog(kafka.Error(err, "TaskDomain.FindProjectIdByTaskId.taskRepo.FindTaskById", kafka.FieldMap{
			"taskId": taskId,
		}))
		return 0, false, model.DBError
	}
	if task == nil {
		return 0, false, nil
	}
	return task.ProjectCode, true, nil
}
