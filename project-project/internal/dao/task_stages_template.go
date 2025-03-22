package dao

import (
	"context"
	"project-project/internal/data"
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
func (t *TaskStagesTemplateDao) FindInProTemIds(ctx context.Context, ids []int) ([]data.MsTaskStagesTemplate, error) {
	var tsts []data.MsTaskStagesTemplate
	session := t.conn.Session(ctx)
	err := session.Where("project_template_code in ?", ids).Find(&tsts).Error
	return tsts, err
}

func (t *TaskStagesTemplateDao) FindByProjectTemplateId(ctx context.Context, projectTemplateCode int) (list []*data.MsTaskStagesTemplate, err error) {
	session := t.conn.Session(ctx)
	err = session.
		Model(&data.MsTaskStagesTemplate{}).
		Where("project_template_code=?", projectTemplateCode).
		Order("sort desc,id asc").
		Find(&list).
		Error
	return
}
