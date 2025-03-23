package domain

import (
	"context"
	"project-common/errs"
	"project-project/internal/dao"
	"project-project/internal/data"
	"project-project/internal/repo"
	"project-project/pkg/model"
	"time"
)

type DepartmentDomain struct {
	departmentRepo repo.DepartmentRepo
}

// FindDepartmentById 根据id查询部门信息
func (d *DepartmentDomain) FindDepartmentById(id int64) (*data.Department, *errs.BError) {
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 调用repo的FindDepartmentById方法查询部门信息。
	dp, err := d.departmentRepo.FindDepartmentById(c, id)
	if err != nil {
		return nil, model.DBError
	}
	return dp, nil
}

// List 部门列表
func (d *DepartmentDomain) List(organizationCode int64, parentDepartmentCode int64, page int64, size int64) ([]*data.DepartmentDisplay, int64, *errs.BError) {
	// 调用repo的ListDepartment方法查询部门列表。
	list, total, err := d.departmentRepo.ListDepartment(organizationCode, parentDepartmentCode, page, size)
	if err != nil {
		return nil, 0, model.DBError
	}
	var dList []*data.DepartmentDisplay
	for _, v := range list {
		dList = append(dList, v.ToDisplay())
	}
	return dList, total, nil
}

// Save 新增部门
func (d *DepartmentDomain) Save(
	organizationCode int64,
	departmentCode int64,
	parentDepartmentCode int64,
	name string) (*data.DepartmentDisplay, *errs.BError) {

	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 调用repo的FindDepartment方法查询部门信息。
	dpm, err := d.departmentRepo.FindDepartment(c, organizationCode, parentDepartmentCode, name)
	if err != nil {
		return nil, model.DBError
	}
	if dpm == nil {
		dpm = &data.Department{
			Name:             name,
			OrganizationCode: organizationCode,
			CreateTime:       time.Now().UnixMilli(),
		}
		if parentDepartmentCode > 0 {
			dpm.Pcode = parentDepartmentCode
		}
		// 调用repo的Save方法保存部门信息。
		err := d.departmentRepo.Save(dpm)
		if err != nil {
			return nil, model.DBError
		}
		return dpm.ToDisplay(), nil
	}
	return dpm.ToDisplay(), nil
}

func NewDepartmentDomain() *DepartmentDomain {
	return &DepartmentDomain{
		departmentRepo: dao.NewDepartmentDao(),
	}
}
