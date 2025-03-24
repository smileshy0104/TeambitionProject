package repo

import (
	"context"
	"project-project/internal/data"
)

type MenuRepo interface {
	// FindMenus 查询菜单
	FindMenus(ctx context.Context) ([]*data.ProjectMenu, error)
}
