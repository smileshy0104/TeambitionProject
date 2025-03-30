package data

import "github.com/jinzhu/copier"

// ProjectMenu 项目菜单
type ProjectMenu struct {
	Id         int64
	Pid        int64
	Title      string
	Icon       string
	Url        string
	FilePath   string
	Params     string
	Node       string
	Sort       int
	Status     int
	CreateBy   int64
	IsInner    int
	Values     string
	ShowSlider int
}

func (*ProjectMenu) TableName() string {
	return "project_menu"
}

// ProjectMenuChild 项目菜单
type ProjectMenuChild struct {
	ProjectMenu                     // 嵌入 ProjectMenu
	Children    []*ProjectMenuChild // 子菜单
}

// CovertChild 将 ProjectMenu 列表转换为 ProjectMenuChild 列表。
// 这个函数首先复制输入的 ProjectMenu 列表到 ProjectMenuChild 列表格式，然后通过调用 toChild 函数来构建每个菜单项的子菜单项。
// 参数 pms 是一个指向 ProjectMenu 列表的指针。
// 返回值是一个指向 ProjectMenuChild 列表的指针，表示转换后的带有子菜单结构的菜单列表。
func CovertChild(pms []*ProjectMenu) []*ProjectMenuChild {
	var pmcs []*ProjectMenuChild
	copier.Copy(&pmcs, pms)
	var childPmcs []*ProjectMenuChild
	// 遍历 pmcs，将每个菜单项的 Pid 为 0 的菜单项添加到 childPmcs 中。
	for _, v := range pmcs {
		if v.Pid == 0 {
			pmc := &ProjectMenuChild{}
			copier.Copy(pmc, v)
			childPmcs = append(childPmcs, pmc)
		}
	}
	// 遍历 childPmcs，为每个菜单项的子菜单项调用 toChild 函数。
	toChild(childPmcs, pmcs)
	return childPmcs
}

// toChild 是一个递归函数，用于构建 ProjectMenuChild 实例的子菜单结构。
// 这个函数通过比较菜单项的父 ID（Pid）和子菜单项的 ID 来构建树形结构的菜单。
// 参数 childPmcs 是当前层级的子菜单列表，pmcs 是所有菜单项的列表。
func toChild(childPmcs []*ProjectMenuChild, pmcs []*ProjectMenuChild) {
	// 遍历 childPmcs，为每个菜单项的子菜单项调用 toChild 函数。
	for _, pmc := range childPmcs {
		for _, pm := range pmcs {
			if pmc.Id == pm.Pid {
				child := &ProjectMenuChild{}
				copier.Copy(child, pm)
				pmc.Children = append(pmc.Children, child)
			}
		}
		// 递归调用自身，为每个子菜单项的子菜单项调用 toChild 函数。
		toChild(pmc.Children, pmcs)
	}
}

func getFullUrl(url string, params string, values string) string {
	if (params != "" && values != "") || values != "" {
		return url + "/" + values
	}
	return url
}

func getInnerText(inner int) string {
	if inner == 0 {
		return "导航"
	}
	if inner == 1 {
		return "内页"
	}
	return ""
}

func getStatus(status int) string {
	if status == 0 {
		return "禁用"
	}
	if status == 1 {
		return "使用中"
	}
	return ""
}
