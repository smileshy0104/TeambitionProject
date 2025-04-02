package domain

import (
	"context"
	"project-common/errs"
	"project-project/internal/dao"
	"project-project/internal/data"
	"project-project/internal/repo"
	"project-project/pkg/model"
	"time"
)

type MenuDomain struct {
	menuRepo repo.MenuRepo
}

func (d *MenuDomain) MenuTreeList() ([]*data.ProjectMenuChild, *errs.BError) {
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	menus, err := d.menuRepo.FindMenus(c)
	if err != nil {
		return nil, model.DBError
	}
	menuChildren := data.CovertChild(menus)
	return menuChildren, nil
}

func NewMenuDomain() *MenuDomain {
	return &MenuDomain{
		menuRepo: dao.NewMenuDao(),
	}
}
