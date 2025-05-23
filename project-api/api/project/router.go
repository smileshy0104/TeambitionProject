package project

import (
	"github.com/gin-gonic/gin"
	"log"
	"project-api/api/midd"
	"project-api/router"
)

type RouterProject struct {
}

func init() {
	log.Println("init project router")
	ru := &RouterProject{}
	router.Register(ru)
}

func (*RouterProject) Route(r *gin.Engine) {
	//初始化grpc的客户端连接
	InitRpcProjectClient()
	h := New()
	// 定义对应的路由组规则
	group := r.Group("/project")
	// 使用TokenVerify中间件对项目列表的API进行身份验证
	group.Use(midd.TokenVerify())
	//group.Use(Auth())
	//group.Use(ProjectAuth())
	group.POST("/index", h.index)                                     // Index 获取项目的菜单列表
	group.POST("/project/selfList", h.myProjectList)                  // myProjectList 获取用户自身项目列表请求
	group.POST("/project", h.myProjectList)                           // myProjectList 获取项目列表请求
	group.POST("/project_template", h.projectTemplate)                // projectTemplate 获取项目模板列表请求
	group.POST("/project/save", h.projectSave)                        // projectSave 保存项目请求
	group.POST("/project/read", h.readProject)                        // readProject 读取项目请求
	group.POST("/project/recycle", h.recycleProject)                  // recycleProject 删除项目请求
	group.POST("/project/recovery", h.recoveryProject)                // recoveryProject 恢复项目请求
	group.POST("/project_collect/collect", h.collectProject)          // collectProject 收藏项目请求
	group.POST("/project/edit", h.editProject)                        // editProject 编辑项目请求
	group.POST("/project/getLogBySelfProject", h.getLogBySelfProject) // getLogBySelfProject 获取项目日志列表请求

	t := NewTask()
	group.POST("/task_stages", t.taskStages)                 // taskStages 查找对应的任务阶段列表。
	group.POST("/project_member/index", t.memberProjectList) // memberProjectList 查询项目成员列表。
	group.POST("/task_stages/tasks", t.taskList)             // taskList 查找对应的任务列表。
	group.POST("/task/save", t.saveTask)                     // saveTask 保存任务。
	group.POST("/task/edit", t.editTask)                     // editTask 修改任务。
	group.POST("/task/sort", t.taskSort)                     // taskSort 任务排序。
	group.POST("/task/selfList", t.myTaskList)               // myTaskList 获取用户自身任务列表请求

	group.POST("/task/read", t.readTask)                      // readTask 读取任务请求
	group.POST("/task_member", t.listTaskMember)              // listTaskMember 查询任务成员列表。
	group.POST("/task/taskLog", t.taskLog)                    // taskLog 查询任务日志列表。
	group.POST("/task/_taskWorkTimeList", t.taskWorkTimeList) // taskWorkTimeList 查询任务工时列表。
	group.POST("/task/saveTaskWorkTime", t.saveTaskWorkTime)  // saveTaskWorkTime 保存任务工时。
	group.POST("/file/uploadFiles", t.uploadFiles)            // uploadFiles 上传文件。
	group.POST("/task/taskSources", t.taskSources)            // taskSources 查询任务来源列表。
	group.POST("/task/createComment", t.createComment)        // createComment 创建评论。

	a := NewAccount()
	group.POST("/account", a.account) //account 获取用户信息。
	d := NewDepartment()
	group.POST("/department", d.department)
	group.POST("/department/save", d.save)
	group.POST("/department/read", d.read)
	auth := NewAuth()
	group.POST("/auth", auth.authList)
	group.POST("/auth/apply", auth.apply)
	menu := NewMenu()
	group.POST("/menu/menu", menu.menuList)
}
