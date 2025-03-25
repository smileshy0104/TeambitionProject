package project

import (
	"github.com/gin-gonic/gin"
	"net/http"
	common "project-common"
	"project-common/errs"
	"strings"
)

var ignores = []string{
	"project/login/register",
	"project/login",
	"project/login/getCaptcha",
	"project/organization",
	"project/auth/apply"}

// Auth 返回一个中间件函数，用于验证请求的授权状态
// 如果请求的URI在忽略列表中，或者匹配到用户的授权节点，则放行请求
// 否则，返回403错误，表示无权限操作
func Auth() func(*gin.Context) {
	return func(c *gin.Context) {
		result := &common.Result{}
		uri := c.Request.RequestURI

		// 检查请求URI是否在忽略列表中
		for _, v := range ignores {
			if strings.Contains(uri, v) {
				c.Next()
				return
			}
		}

		// 判断此uri是否在用户的授权列表中
		a := NewAuth()
		nodes, err := a.GetAuthNodes(c)
		if err != nil {
			code, msg := errs.ParseGrpcError(err)
			c.JSON(http.StatusOK, result.Fail(code, msg))
			c.Abort()
			return
		}

		// 检查请求URI是否匹配用户的授权节点
		for _, v := range nodes {
			if strings.Contains(uri, v) {
				c.Next()
				return
			}
		}

		// 如果请求URI不在授权列表中，返回403错误
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作"))
		c.Abort()
		return
	}
}
