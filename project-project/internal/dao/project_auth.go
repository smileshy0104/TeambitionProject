package dao

import (
	"context"
	"project-project/internal/data"
	"project-project/internal/database/gorms"
)

type ProjectAuthDao struct {
	conn *gorms.GormConn
}

// FindAuthListPage 查询项目权限列表分页
func (p *ProjectAuthDao) FindAuthListPage(ctx context.Context, orgCode int64, page int64, pageSize int64) (list []*data.ProjectAuth, total int64, err error) {
	session := p.conn.Session(ctx)
	err = session.Model(&data.ProjectAuth{}).
		Where("organization_code=?", orgCode).
		Limit(int(pageSize)).
		Offset(int((page - 1) * pageSize)).
		Find(&list).Error
	err = session.Model(&data.ProjectAuth{}).
		Where("organization_code=?", orgCode).
		Count(&total).Error
	return
}

// FindAuthList 查询项目权限列表
func (p *ProjectAuthDao) FindAuthList(ctx context.Context, orgCode int64) (list []*data.ProjectAuth, err error) {
	session := p.conn.Session(ctx)
	err = session.Model(&data.ProjectAuth{}).Where("organization_code=?", orgCode).Find(&list).Error
	return
}

func NewProjectAuthDao() *ProjectAuthDao {
	return &ProjectAuthDao{
		conn: gorms.New(),
	}
}
