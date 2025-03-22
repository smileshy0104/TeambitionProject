package dao

import (
	"context"
	"project-project/internal/data/task"
	"project-project/internal/database/gorms"
)

type TaskStagesTemplateDao struct {
	conn *gorms.GormConn
}

func NewTaskStagesTemplateDao() *TaskStagesTemplateDao {
	return &TaskStagesTemplateDao{
		conn: gorms.New(),
	}
}

// FindInProTemIds 根据项目模板id查询
func (t *TaskStagesTemplateDao) FindInProTemIds(ctx context.Context, ids []int) ([]task.MsTaskStagesTemplate, error) {
	var tsts []task.MsTaskStagesTemplate
	session := t.conn.Session(ctx)
	err := session.Where("project_template_code in ?", ids).Find(&tsts).Error
	return tsts, err
}
