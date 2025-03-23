package dao

import (
	"context"
	"gorm.io/gorm"
	"project-project/internal/data/task"
	"project-project/internal/database"
	"project-project/internal/database/gorms"
)

type TaskDao struct {
	conn *gorms.GormConn
}

// FindTaskByIds 根据任务ID查询任务
func (t *TaskDao) FindTaskByIds(background context.Context, taskIdList []int64) (list []*task.Task, err error) {
	session := t.conn.Session(background)
	err = session.Model(&task.Task{}).Where("id in (?)", taskIdList).Find(&list).Error
	return
}

// FindTaskMemberPage 根据任务ID分页查询任务成员
func (t *TaskDao) FindTaskMemberPage(ctx context.Context, taskCode int64, page int64, size int64) (list []*task.TaskMember, total int64, err error) {
	session := t.conn.Session(ctx)
	offset := (page - 1) * size
	err = session.Model(&task.TaskMember{}).
		Where("task_code=?", taskCode).
		Limit(int(size)).Offset(int(offset)).
		Find(&list).Error
	err = session.Model(&task.TaskMember{}).
		Where("task_code=?", taskCode).
		Count(&total).Error
	return
}

// FindTaskByAssignTo 根据成员ID查询任务
func (t *TaskDao) FindTaskByAssignTo(ctx context.Context, memberId int64, done int, page int64, size int64) (tsList []*task.Task, total int64, err error) {
	session := t.conn.Session(ctx)
	offset := (page - 1) * size
	err = session.Model(&task.Task{}).Where("assign_to=? and deleted=0 and done=?", memberId, done).Limit(int(size)).Offset(int(offset)).Find(&tsList).Error
	err = session.Model(&task.Task{}).Where("assign_to=? and deleted=0 and done=?", memberId, done).Count(&total).Error
	return
}

// FindTaskByMemberCode 根据成员ID查询任务
func (t *TaskDao) FindTaskByMemberCode(ctx context.Context, memberId int64, done int, page int64, size int64) (tList []*task.Task, total int64, err error) {
	session := t.conn.Session(ctx)
	offset := (page - 1) * size
	sql := "select a.* from ms_task a,ms_task_member b where a.id=b.task_code and member_code=? and a.deleted=0 and a.done=? limit ?,?"
	raw := session.Model(&task.Task{}).Raw(sql, memberId, done, offset, size)
	err = raw.Scan(&tList).Error
	if err != nil {
		return nil, 0, err
	}
	sqlCount := "select count(*) from ms_task a,ms_task_member b where a.id=b.task_code and member_code=? and a.deleted=0 and a.done=?"
	rawCount := session.Model(&task.Task{}).Raw(sqlCount, memberId, done)
	err = rawCount.Scan(&total).Error
	return
}

// FindTaskByCreateBy 根据创建人ID查询任务
func (t *TaskDao) FindTaskByCreateBy(ctx context.Context, memberId int64, done int, page int64, size int64) (tList []*task.Task, total int64, err error) {
	session := t.conn.Session(ctx)
	offset := (page - 1) * size
	err = session.Model(&task.Task{}).Where("create_by=? and deleted=0 and done=?", memberId, done).Limit(int(size)).Offset(int(offset)).Find(&tList).Error
	err = session.Model(&task.Task{}).Where("create_by=? and deleted=0 and done=?", memberId, done).Count(&total).Error
	return
}

// FindTaskByStageCodeLtSort 根据阶段ID和排序查询任务
func (t *TaskDao) FindTaskByStageCodeLtSort(ctx context.Context, stageCode int, sort int) (ts *task.Task, err error) {
	session := t.conn.Session(ctx)
	err = session.Where("stage_code=? and sort < ?", stageCode, sort).Order("sort desc").Limit(1).Find(&ts).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return
}

// UpdateTaskSort 更新任务排序
func (t *TaskDao) UpdateTaskSort(ctx context.Context, conn database.DbConn, ts *task.Task) error {
	t.conn = conn.(*gorms.GormConn)
	err := t.conn.Tx(ctx).
		Where("id=?", ts.Id).
		Select("sort", "stage_code").
		Updates(&ts).
		Error
	return err
}

// FindTaskById 根据任务ID查询任务
func (t *TaskDao) FindTaskById(ctx context.Context, taskCode int64) (ts *task.Task, err error) {
	session := t.conn.Session(ctx)
	err = session.Where("id=?", taskCode).Find(&ts).Error
	return
}

// SaveTask 保存任务
func (t *TaskDao) SaveTask(ctx context.Context, conn database.DbConn, ts *task.Task) error {
	t.conn = conn.(*gorms.GormConn)
	err := t.conn.Tx(ctx).Save(&ts).Error
	return err
}

// SaveTaskMember 保存任务成员
func (t *TaskDao) SaveTaskMember(ctx context.Context, conn database.DbConn, tm *task.TaskMember) error {
	t.conn = conn.(*gorms.GormConn)
	err := t.conn.Tx(ctx).Save(&tm).Error
	return err
}

// FindTaskSort 根据项目ID和阶段ID查询任务排序
func (t *TaskDao) FindTaskSort(ctx context.Context, projectCode int64, stageCode int64) (v *int, err error) {
	session := t.conn.Session(ctx)
	//select * from
	err = session.Model(&task.Task{}).
		Where("project_code=? and stage_code=?", projectCode, stageCode).
		Select("max(sort)").Scan(&v).Error
	return
}

// FindTaskMaxIdNum 根据项目ID查询最大ID
func (t *TaskDao) FindTaskMaxIdNum(ctx context.Context, projectCode int64) (v *int, err error) {
	session := t.conn.Session(ctx)
	//select * from
	err = session.Model(&task.Task{}).
		Where("project_code=?", projectCode).
		Select("max(id_num)").Scan(&v).Error
	return
}

// FindTaskMemberByTaskId 根据任务ID和成员ID查询任务成员
func (t *TaskDao) FindTaskMemberByTaskId(ctx context.Context, taskCode int64, memberId int64) (task *task.TaskMember, err error) {
	err = t.conn.Session(ctx).
		Where("task_code=? and member_code=?", taskCode, memberId).
		Limit(1).
		Find(&task).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return
}

// FindTaskByStageCode 根据StageCode阶段查询任务
func (t *TaskDao) FindTaskByStageCode(ctx context.Context, stageCode int) (list []*task.Task, err error) {
	//select * from ms_task where stage_code=77 and deleted =0 order by sort asc
	session := t.conn.Session(ctx)
	err = session.Model(&task.Task{}).
		Where("stage_code=? and deleted =0", stageCode).
		Order("sort asc").
		Find(&list).Error
	return
}

func NewTaskDao() *TaskDao {
	return &TaskDao{
		conn: gorms.New(),
	}
}
