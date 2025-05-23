package domain

import (
	"context"
	"go.uber.org/zap"
	"project-common/errs"
	"project-project/internal/dao"
	"project-project/internal/data"
	"project-project/internal/database"
	"project-project/internal/repo"
	"project-project/pkg/model"
	"strconv"
	"time"
)

type ProjectAuthDomain struct {
	projectAuthRepo       repo.ProjectAuthRepo
	userRpcDomain         *UserRpcDomain
	projectNodeDomain     *ProjectNodeDomain
	projectAuthNodeDomain *ProjectAuthNodeDomain
	accountDomain         *AccountDomain
}

// AuthList 查询权限列表
func (d *ProjectAuthDomain) AuthList(orgCode int64) ([]*data.ProjectAuthDisplay, *errs.BError) {
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 查询权限列表
	list, err := d.projectAuthRepo.FindAuthList(c, orgCode)
	if err != nil {
		zap.L().Error("project AuthList projectAuthRepo.FindAuthList error", zap.Error(err))
		return nil, model.DBError
	}
	var pdList []*data.ProjectAuthDisplay
	for _, v := range list {
		display := v.ToDisplay()
		pdList = append(pdList, display)
	}
	return pdList, nil
}

// AuthListPage 分页查询
func (d *ProjectAuthDomain) AuthListPage(orgCode int64, page int64, pageSize int64) ([]*data.ProjectAuthDisplay, int64, *errs.BError) {
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	list, total, err := d.projectAuthRepo.FindAuthListPage(c, orgCode, page, pageSize)
	if err != nil {
		zap.L().Error("project AuthList projectAuthRepo.FindAuthList error", zap.Error(err))
		return nil, 0, model.DBError
	}
	var pdList []*data.ProjectAuthDisplay
	for _, v := range list {
		display := v.ToDisplay()
		pdList = append(pdList, display)
	}
	return pdList, total, nil
}

func (d *ProjectAuthDomain) AllNodeAndAuth(authId int64) ([]*data.ProjectNodeAuthTree, []string, *errs.BError) {
	nodeList, err := d.projectNodeDomain.NodeList()
	if err != nil {
		return nil, nil, err
	}
	checkedList, err := d.projectAuthNodeDomain.AuthNodeList(authId)
	if err != nil {
		return nil, nil, err
	}
	list := data.ToAuthNodeTreeList(nodeList, checkedList)
	return list, checkedList, nil
}

func (d *ProjectAuthDomain) Save(conn database.DbConn, authId int64, nodes []string) *errs.BError {
	err := d.projectAuthNodeDomain.Save(conn, authId, nodes)
	if err != nil {
		return err
	}
	return nil
}

func (d *ProjectAuthDomain) AuthNodes(memberId int64) ([]string, *errs.BError) {
	account, err := d.accountDomain.FindAccount(memberId)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, model.ParamsError
	}
	//c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	authorize := account.Authorize
	authId, _ := strconv.ParseInt(authorize, 10, 64)
	authNodeList, dbErr := d.projectAuthNodeDomain.AuthNodeList(authId)
	if dbErr != nil {
		return nil, model.DBError
	}
	return authNodeList, nil
}

func NewProjectAuthDomain() *ProjectAuthDomain {
	return &ProjectAuthDomain{
		projectAuthRepo:       dao.NewProjectAuthDao(),
		userRpcDomain:         NewUserRpcDomain(),
		projectNodeDomain:     NewProjectNodeDomain(),
		projectAuthNodeDomain: NewProjectAuthNodeDomain(),
		accountDomain:         NewAccountDomain(),
	}
}
