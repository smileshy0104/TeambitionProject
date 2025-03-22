package dao

import (
	"context"
	"project-project/internal/data/pro"
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

// FindProjectByMemId 根据成员ID查找项目信息。
// 该方法使用SQL查询来获取特定成员参与的所有项目，并计算总项目数。
// 参数:
//
//	ctx - 上下文，用于取消请求和传递请求级值。
//	memId - 成员ID，用于查找该成员参与的项目。
//	page - 页码，用于分页查询。
//	size - 每页大小，用于限制查询结果的数量。
//
// 返回值:
//
//	[]*pro.ProjectAndMember - 项目列表，每个项目都包含成员信息。
//	int64 - 成员参与的总项目数。
//	error - 错误信息，如果查询过程中发生错误。
func (p ProjectDao) FindProjectByMemId(ctx context.Context, memId int64, page int64, size int64) ([]*pro.ProjectAndMember, int64, error) {
	// 初始化项目列表。
	var pms []*pro.ProjectAndMember

	// 获取数据库会话。
	session := p.conn.Session(ctx)

	// 计算查询的起始索引。
	index := (page - 1) * size

	// 执行SQL查询，获取项目列表。
	raw := session.Raw("select * from ms_project a, ms_project_member b where a.id = b.project_code and b.member_code=? limit ?,?", memId, index, size)
	raw.Scan(&pms)

	// 初始化总项目数。
	var total int64

	// 计算成员参与的总项目数。
	err := session.Model(&pro.ProjectMember{}).Where("member_code=?", memId).Count(&total).Error

	// 返回项目列表、总项目数和可能的错误。
	return pms, total, err
}
