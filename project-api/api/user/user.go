package user

import (
	"github.com/gin-gonic/gin"
	common "project-common"
)

type HandlerUser struct {
}

func New() *HandlerUser {
	return &HandlerUser{}
}

func (*HandlerUser) getCaptcha(ctx *gin.Context) {
	result := &common.Result{}
	ctx.JSON(200, result)
}
