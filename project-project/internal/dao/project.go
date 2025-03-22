package dao

import (
	"context"
	"fmt"
	"project-project/internal/data/pro"
	"project-project/internal/database"
	"project-project/internal/database/gorms"
)

type ProjectDao struct {
	conn *gorms.GormConn
}

func NewProjectDao() *ProjectDao {
	return &ProjectDao{
		conn: gorms.New(),
	}
}

// UpdateProject 更新项目信息
func (p *ProjectDao) UpdateProject(ctx context.Context, proj *pro.Project) error {
	return p.conn.Session(ctx).Updates(&proj).Error
}

// DeleteProjectCollect 删除项目收藏
func (p *ProjectDao) DeleteProjectCollect(ctx context.Context, memId int64, projectCode int64) error {
	return p.conn.Session(ctx).Where("member_code=? and project_code=?", memId, projectCode).Delete(&pro.ProjectCollection{}).Error
}

// SaveProjectCollect 保存项目收藏
func (p *ProjectDao) SaveProjectCollect(ctx context.Context, pc *pro.ProjectCollection) error {
	return p.conn.Session(ctx).Save(&pc).Error
}

// UpdateDeletedProject 更新项目删除状态
func (p *ProjectDao) UpdateDeletedProject(ctx context.Context, code int64, deleted bool) error {
	session := p.conn.Session(ctx)
	var err error
	if deleted {
		err = session.Model(&pro.Project{}).Where("id=?", code).Update("deleted", 1).Error
	} else {
		err = session.Model(&pro.Project{}).Where("id=?", code).Update("deleted", 0).Error
	}
	return err
}

// FindProjectByPIdAndMemId 根据项目id和成员id查询项目信息
func (p *ProjectDao) FindProjectByPIdAndMemId(ctx context.Context, projectCode int64, memberId int64) (*pro.ProjectAndMember, error) {
	var pms *pro.ProjectAndMember
	session := p.conn.Session(ctx)
	sql := fmt.Sprintf("select a.*,b.project_code,b.member_code,b.join_time,b.is_owner,b.authorize from ms_project a, ms_project_member b where a.id = b.project_code and b.member_code=? and b.project_code=? limit 1")
	raw := session.Raw(sql, memberId, projectCode)
	err := raw.Scan(&pms).Error
	return pms, err
}

// FindCollectByPidAndMemId 根据项目id和成员id查询是否收藏了项目
func (p *ProjectDao) FindCollectByPidAndMemId(ctx context.Context, projectCode int64, memberId int64) (bool, error) {
	var count int64
	session := p.conn.Session(ctx)
	sql := fmt.Sprintf("select count(*) from ms_project_collection where member_code=? and project_code=?")
	raw := session.Raw(sql, memberId, projectCode)
	err := raw.Scan(&count).Error
	return count > 0, err
}

// SaveProject 保存项目
func (p *ProjectDao) SaveProject(conn database.DbConn, ctx context.Context, pr *pro.Project) error {
	p.conn = conn.(*gorms.GormConn)
	return p.conn.Tx(ctx).Save(&pr).Error
}

// SaveProjectMember 保存项目成员
func (p *ProjectDao) SaveProjectMember(conn database.DbConn, ctx context.Context, pm *pro.ProjectMember) error {
	p.conn = conn.(*gorms.GormConn)
	return p.conn.Tx(ctx).Save(&pm).Error
}

// FindCollectProjectByMemId 根据成员id查询被收藏的项目
func (p ProjectDao) FindCollectProjectByMemId(ctx context.Context, memberId int64, page int64, size int64) ([]*pro.ProjectAndMember, int64, error) {
	var pms []*pro.ProjectAndMember
	session := p.conn.Session(ctx)
	index := (page - 1) * size
	sql := fmt.Sprintf("select * from ms_project where id in (select project_code from ms_project_collection where member_code=?) order by sort limit ?,?")
	raw := session.Raw(sql, memberId, index, size)
	raw.Scan(&pms)
	var total int64
	query := fmt.Sprintf("member_code=?")
	err := session.Model(&pro.ProjectCollection{}).Where(query, memberId).Count(&total).Error
	return pms, total, err
}

// FindProjectByMemId 根据成员id查询项目
func (p ProjectDao) FindProjectByMemId(ctx context.Context, memId int64, condition string, page int64, size int64) ([]*pro.ProjectAndMember, int64, error) {
	var pms []*pro.ProjectAndMember
	session := p.conn.Session(ctx)
	index := (page - 1) * size
	sql := fmt.Sprintf("select * from ms_project a, ms_project_member b where a.id = b.project_code and b.member_code=?  %s order by sort limit ?,?", condition)
	raw := session.Raw(sql, memId, index, size)
	raw.Scan(&pms)
	var total int64
	query := fmt.Sprintf("select count(*) from ms_project a, ms_project_member b where a.id = b.project_code and b.member_code=?  %s", condition)
	tx := session.Raw(query, memId)
	err := tx.Scan(&total).Error
	return pms, total, err
}
