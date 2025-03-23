package repo

import (
	"context"
	"project-project/internal/data"
)

type DepartmentRepo interface {
	// 根据id查询部门
	FindDepartmentById(ctx context.Context, id int64) (*data.Department, error)
	// 根据条件查询部门
	FindDepartment(ctx context.Context, organizationCode int64, parentDepartmentCode int64, name string) (*data.Department, error)
	// 保存部门
	Save(dpm *data.Department) error
	// 查询部门列表
	ListDepartment(organizationCode int64, parentDepartmentCode int64, page int64, size int64) (list []*data.Department, total int64, err error)
}
