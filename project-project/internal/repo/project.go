package repo

import (
	"context"
	"project-project/internal/data/pro"
)

type ProjectRepo interface {
	// FindProjectByMemId 根据成员id查询项目
	FindProjectByMemId(ctx context.Context, memId int64, page int64, size int64) ([]*pro.ProjectAndMember, int64, error)
}
