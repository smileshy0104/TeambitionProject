package repo

import (
	"context"
	"project-project/internal/data"
)

type FileRepo interface {
	// Save 保存文件信息
	Save(ctx context.Context, file *data.File) error
	// FindByIds 根据id查询文件信息
	FindByIds(background context.Context, ids []int64) (list []*data.File, err error)
}
