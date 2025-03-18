package user

import (
	"github.com/gin-gonic/gin"
	"log"
	"project-api/router"
)

type RouterUser struct {
}

func init() {
	log.Println("init user router")
	ru := &RouterUser{}
	router.Register(ru)
}

func (*RouterUser) Route(r *gin.Engine) {
	h := New()
	r.POST("/project/login/getCaptcha", h.getCaptcha)
}
