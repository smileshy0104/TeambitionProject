package pro

import (
	"project-common/encrypts"
	"project-common/tms"
	"project-project/internal/data/task"
	"project-project/pkg/model"
)

// Project 项目
type Project struct {
	Id                 int64
	Cover              string
	Name               string
	Description        string
	AccessControlType  int
	WhiteList          string
	Sort               int
	Deleted            int
	TemplateCode       int
	Schedule           float64
	CreateTime         int64
	OrganizationCode   int64
	DeletedTime        string
	Private            int
	Prefix             string
	OpenPrefix         int
	Archive            int
	ArchiveTime        int64
	OpenBeginTime      int
	OpenTaskPrivate    int
	TaskBoardTheme     string
	BeginTime          int64
	EndTime            int64
	AutoUpdateSchedule int
}

func (*Project) TableName() string {
	return "project"
}

// ProjectMember 项目成员
type ProjectMember struct {
	Id          int64
	ProjectCode int64
	MemberCode  int64
	JoinTime    int64
	IsOwner     int64
	Authorize   string
}

func (*ProjectMember) TableName() string {
	return "project_member"
}

type ProjectCollection struct {
	Id          int64
	ProjectCode int64
	MemberCode  int64
	CreateTime  int64
}

func (*ProjectCollection) TableName() string {
	return "project_collection"
}

// ProjectAndMember 项目和成员
type ProjectAndMember struct {
	Project
	ProjectCode int64
	MemberCode  int64
	JoinTime    int64
	IsOwner     int64
	Authorize   string
	OwnerName   string
	Collected   int
}

func (m *ProjectAndMember) GetAccessControlType() string {
	if m.AccessControlType == 0 {
		return "open"
	}
	if m.AccessControlType == 1 {
		return "private"
	}
	if m.AccessControlType == 2 {
		return "custom"
	}
	return ""
}

func (m *Project) GetAccessControlType() string {
	if m.AccessControlType == 0 {
		return "open"
	}
	if m.AccessControlType == 1 {
		return "private"
	}
	if m.AccessControlType == 2 {
		return "custom"
	}
	return ""
}

func ToMap(orgs []*ProjectAndMember) map[int64]*ProjectAndMember {
	m := make(map[int64]*ProjectAndMember)
	for _, v := range orgs {
		m[v.Id] = v
	}
	return m
}

type ProjectTemplate struct {
	Id               int
	Name             string
	Description      string
	Sort             int
	CreateTime       int64
	OrganizationCode int64
	Cover            string
	MemberCode       int64
	IsSystem         int
}

func (*ProjectTemplate) TableName() string {
	return "project_template"
}

type ProjectTemplateAll struct {
	Id               int
	Name             string
	Description      string
	Sort             int
	CreateTime       string
	OrganizationCode string
	Cover            string
	MemberCode       string
	IsSystem         int
	TaskStages       []*task.TaskStagesOnlyName
	Code             string
}

// Convert 方法用于将 ProjectTemplate 类型的实例转换为 ProjectTemplateAll 类型的实例。
// 这个方法主要负责加密一些敏感字段，并组装一个新的结构体实例。
// 参数 taskStages 是一个指向任务阶段名称列表的指针，表示项目模板中的各个任务阶段。
func (pt ProjectTemplate) Convert(taskStages []*task.TaskStagesOnlyName) *ProjectTemplateAll {
	// 加密组织代码
	organizationCode, _ := encrypts.EncryptInt64(pt.OrganizationCode, model.AESKey)
	// 加密成员代码
	memberCode, _ := encrypts.EncryptInt64(pt.MemberCode, model.AESKey)
	// 加密项目模板代码（这里使用项目模板的 ID 进行加密）
	code, _ := encrypts.EncryptInt64(int64(pt.Id), model.AESKey)

	// 创建并初始化一个 ProjectTemplateAll 类型的实例
	pta := &ProjectTemplateAll{
		Id:               pt.Id,
		Name:             pt.Name,
		Description:      pt.Description,
		Sort:             pt.Sort,
		CreateTime:       tms.FormatByMill(pt.CreateTime),
		OrganizationCode: organizationCode,
		Cover:            pt.Cover,
		MemberCode:       memberCode,
		IsSystem:         pt.IsSystem,
		TaskStages:       taskStages,
		Code:             code,
	}
	// 返回转换后的 ProjectTemplateAll 实例
	return pta
}

// ToProjectTemplateIds 获取项目模板ids
func ToProjectTemplateIds(pts []ProjectTemplate) []int {
	var ids []int
	for _, v := range pts {
		ids = append(ids, v.Id)
	}
	return ids
}
