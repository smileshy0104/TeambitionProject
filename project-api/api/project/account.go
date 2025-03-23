package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/pkg/model"
	common "project-common"
	"project-common/errs"
	"project-grpc/account"
	"time"
)

type HandlerAccount struct {
}

// account 账户列表
func (a *HandlerAccount) account(c *gin.Context) {
	//接收请求参数  一些参数的校验 可以放在api这里
	result := &common.Result{}
	var req *model.AccountReq
	_ = c.ShouldBind(&req)
	// 获取当前登录用户id
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 调用project模块 查询账户列表
	msg := &account.AccountReqMessage{
		MemberId:         memberId,
		OrganizationCode: c.GetString("organizationCode"),
		Page:             int64(req.Page),
		PageSize:         int64(req.PageSize),
		SearchType:       int32(req.SearchType),
		DepartmentCode:   req.DepartmentCode,
	}
	// 调用AccountServiceClient的Account方法获取账户列表。
	response, err := AccountServiceClient.Account(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	// 拷贝账单列表
	var list []*model.MemberAccount
	copier.Copy(&list, response.AccountList)
	if list == nil {
		list = []*model.MemberAccount{}
	}
	// 获取项目权限列表
	var authList []*model.ProjectAuth
	copier.Copy(&authList, response.AuthList)
	if authList == nil {
		authList = []*model.ProjectAuth{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"total":    response.Total,
		"page":     req.Page,
		"list":     list,
		"authList": authList,
	}))
}

func NewAccount() *HandlerAccount {
	return &HandlerAccount{}
}
