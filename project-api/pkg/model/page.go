package model

import "github.com/gin-gonic/gin"

// Page 分页结构体，用于分页查询
type Page struct {
	// Page 当前页码
	Page int64 `json:"page" form:"page"`
	// PageSize 每页记录数
	PageSize int64 `json:"pageSize" form:"pageSize"`
}

// Bind 从gin的Context中绑定分页参数，设置默认的分页值
// 参数:
//
//	c *gin.Context - Gin的上下文，用于绑定请求参数
func (p *Page) Bind(c *gin.Context) {
	// 尝试从请求中绑定分页参数，错误被忽略
	_ = c.ShouldBind(&p)
	// 如果当前页码未设置，则默认为第1页
	if p.Page == 0 {
		p.Page = 1
	}
	// 如果每页记录数未设置，则默认为10条
	if p.PageSize == 0 {
		p.PageSize = 10
	}
}
