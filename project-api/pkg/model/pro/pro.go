// Package pro 包含了项目相关的数据结构定义。
package pro

// Project 结构体定义了一个项目的详细信息，包括项目封面、名称、描述等。
type Project struct {
	Cover              string  `json:"cover"`              // 项目封面图片的URL。
	Name               string  `json:"name"`               // 项目名称。
	Description        string  `json:"description"`        // 项目描述。
	AccessControlType  int     `json:"accessControlType"`  // 访问控制类型，决定了谁可以看到或编辑项目。
	WhiteList          string  `json:"whiteList"`          // 白名单，指定可以访问项目的人或组。
	Order              int     `json:"order"`              // 项目的排序顺序。
	Deleted            int     `json:"deleted"`            // 删除标记，表示项目是否已被删除。
	TemplateCode       string  `json:"templateCode"`       // 项目模板代码，标识项目使用的模板。
	Schedule           float64 `json:"schedule"`           // 项目进度的百分比。
	CreateTime         string  `json:"createTime"`         // 项目创建的时间。
	OrganizationCode   int64   `json:"organizationCode"`   // 组织代码，标识项目所属的组织。
	DeletedTime        string  `json:"deletedTime"`        // 项目被删除的时间。
	Private            int     `json:"private"`            // 私有标记，表示项目是否是私有的。
	Prefix             string  `json:"prefix"`             // 项目前缀，用于标识项目。
	OpenPrefix         int     `json:"openPrefix"`         // 开放前缀标记，表示是否公开项目前缀。
	Archive            int     `json:"archive"`            // 归档标记，表示项目是否已被归档。
	ArchiveTime        int64   `json:"archiveTime"`        // 项目被归档的时间。
	OpenBeginTime      int     `json:"openBeginTime"`      // 公开开始时间，项目对成员可见的开始时间。
	OpenTaskPrivate    int     `json:"openTaskPrivate"`    // 公开任务私有标记，表示公开任务是否私有。
	TaskBoardTheme     string  `json:"taskBoardTheme"`     // 任务看板的主题颜色。
	BeginTime          int64   `json:"beginTime"`          // 项目开始的时间。
	EndTime            int64   `json:"endTime"`            // 项目结束的时间。
	AutoUpdateSchedule int     `json:"autoUpdateSchedule"` // 自动更新进度标记，表示项目进度是否自动更新。
	Code               string  `json:"code"`               // 项目代码，项目的唯一标识。
}

// ProjectMember 结构体定义了项目成员的信息，包括成员ID、加入时间等。
type ProjectMember struct {
	Id          int64  `json:"id"`          // 项目成员的ID。
	ProjectCode int64  `json:"projectCode"` // 项目代码，标识成员所属的项目。
	MemberCode  int64  `json:"memberCode"`  // 成员代码，标识项目中的成员。
	JoinTime    int64  `json:"joinTime"`    // 成员加入项目的时间。
	IsOwner     int64  `json:"isOwner"`     // 是否是项目所有者的标记。
	Authorize   string `json:"authorize"`   // 成员的权限，定义了成员在项目中的角色和能力。
}

// ProjectAndMember 结构体结合了项目和成员的信息，用于表示项目及其成员的详细情况。
type ProjectAndMember struct {
	Project            // 嵌入Project结构体，继承项目的所有字段。
	ProjectCode int64  `json:"projectCode"` // 项目代码，用于标识项目。
	MemberCode  int64  `json:"memberCode"`  // 成员代码，用于标识项目中的成员。
	JoinTime    int64  `json:"joinTime"`    // 成员加入项目的时间。
	IsOwner     int64  `json:"isOwner"`     // 是否是项目所有者的标记。
	Authorize   string `json:"authorize"`   // 成员的权限，定义了成员在项目中的角色和能力。
}
