package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/pkg/model"
	common "project-common"
	"project-common/errs"
	"project-grpc/department"
	"time"
)

type HandlerDepartment struct {
}

// department 部门列表
func (d *HandlerDepartment) department(c *gin.Context) {
	result := &common.Result{}
	var req *model.DepartmentReq
	c.ShouldBind(&req)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	msg := &department.DepartmentReqMessage{
		Page:                 req.Page,
		PageSize:             req.PageSize,
		ParentDepartmentCode: req.Pcode,
		OrganizationCode:     c.GetString("organizationCode"),
	}
	// 调用DepartmentServiceClient的List方法获取部门列表。
	listDepartmentMessage, err := DepartmentServiceClient.List(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var list []*model.Department
	copier.Copy(&list, listDepartmentMessage.List)
	if list == nil {
		list = []*model.Department{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"total": listDepartmentMessage.Total,
		"page":  req.Page,
		"list":  list,
	}))
}

// save 添加部门
func (d *HandlerDepartment) save(c *gin.Context) {
	result := &common.Result{}
	var req *model.DepartmentReq
	c.ShouldBind(&req)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	msg := &department.DepartmentReqMessage{
		Name:                 req.Name,
		DepartmentCode:       req.DepartmentCode,
		ParentDepartmentCode: req.ParentDepartmentCode,
		OrganizationCode:     c.GetString("organizationCode"),
	}
	// 调用DepartmentServiceClient的Save方法添加部门。
	departmentMessage, err := DepartmentServiceClient.Save(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var res = &model.Department{}
	copier.Copy(res, departmentMessage)
	c.JSON(http.StatusOK, result.Success(res))
}

// read 读取部门信息
func (d *HandlerDepartment) read(c *gin.Context) {
	result := &common.Result{}
	departmentCode := c.PostForm("departmentCode")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	msg := &department.DepartmentReqMessage{
		DepartmentCode:   departmentCode,
		OrganizationCode: c.GetString("organizationCode"),
	}
	// 调用DepartmentServiceClient的Read方法读取部门信息。
	departmentMessage, err := DepartmentServiceClient.Read(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var res = &model.Department{}
	copier.Copy(res, departmentMessage)
	c.JSON(http.StatusOK, result.Success(res))
}

func NewDepartment() *HandlerDepartment {
	return &HandlerDepartment{}
}
