package dao

import (
	"context"
	"project-project/internal/data"
	"project-project/internal/database/gorms"
)

type FileDao struct {
	conn *gorms.GormConn
}

// FindByIds 根据id查询
func (f *FileDao) FindByIds(ctx context.Context, ids []int64) (list []*data.File, err error) {
	session := f.conn.Session(ctx)
	err = session.Model(&data.File{}).Where("id in (?)", ids).Find(&list).Error
	return
}

// Save 保存
func (f *FileDao) Save(ctx context.Context, file *data.File) error {
	err := f.conn.Session(ctx).Save(&file).Error
	return err
}

func NewFileDao() *FileDao {
	return &FileDao{
		conn: gorms.New(),
	}
}
