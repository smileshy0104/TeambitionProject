package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/pkg/model"
	common "project-common"
	"project-common/errs"
	"project-grpc/menu"
	"time"
)

type HandlerMenu struct {
}

// menuList 处理菜单列表的请求。
// 该函数通过gRPC调用获取菜单列表，并将结果返回给前端。
// 参数:
//
//	c *gin.Context: Gin框架的上下文，用于处理HTTP请求和响应。
func (m HandlerMenu) menuList(c *gin.Context) {
	// 初始化一个通用的结果对象，用于后续构造响应。
	result := &common.Result{}

	// 创建一个带有超时的上下文，以确保gRPC调用不会无限期地等待。
	// 这里设置超时时间为2秒。
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 延迟执行cancel函数，以确保在函数退出时取消上下文。
	defer cancel()

	// 调用gRPC服务，获取菜单列表。
	res, err := MenuServiceClient.MenuList(ctx, &menu.MenuReqMessage{})
	if err != nil {
		// 如果发生错误，解析gRPC错误并返回错误响应。
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	// 初始化菜单列表变量，并将gRPC响应中的列表复制到它。
	var list []*model.Menu
	copier.Copy(&list, res.List)

	// 如果列表为空，确保返回一个空数组而不是nil。
	if list == nil {
		list = []*model.Menu{}
	}

	// 返回成功响应，包含菜单列表。
	c.JSON(http.StatusOK, result.Success(list))
}

func NewMenu() *HandlerMenu {
	return &HandlerMenu{}
}
