package task

type TaskStages struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	ProjectCode int64  `json:"project_code"`
	Sort        int    `json:"sort"`
	Description string `json:"description"`
	CreateTime  int64  `json:"create_time"`
	Deleted     int    `json:"deleted"`
}

func (*TaskStages) TableName() string {
	return "task_stages"
}

// ToTaskStagesMap 将任务阶段的切片转换为字典，其中键为任务阶段的ID。
// 这个函数接收一个指向TaskStages结构体的切片作为参数，
// 并返回一个映射，其中每个元素的键是TaskStages结构体中的Id字段，
// 值是该结构体的指针。
// 这样的转换便于根据任务阶段ID快速查询任务阶段信息。
func ToTaskStagesMap(tsList []*TaskStages) map[int]*TaskStages {
	// 初始化一个空的map用于存储任务阶段信息。
	m := make(map[int]*TaskStages)

	// 遍历任务阶段切片，将每个任务阶段的ID作为键，
	// 任务阶段的指针作为值存入map中。
	for _, v := range tsList {
		m[v.Id] = v
	}

	// 返回构建的map，以便在函数外部使用。
	return m
}
