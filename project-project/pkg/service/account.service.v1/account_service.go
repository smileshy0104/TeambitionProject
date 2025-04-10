package account_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"project-common/business"
	"project-common/errs"
	"project-grpc/account"
	"project-project/internal/dao"
	"project-project/internal/database/tran"
	"project-project/internal/domain"
	"project-project/internal/repo"
)

type AccountService struct {
	account.UnimplementedAccountServiceServer
	cache             repo.Cache
	transaction       tran.Transaction
	accountDomain     *domain.AccountDomain
	projectAuthDomain *domain.ProjectAuthDomain
}

func New() *AccountService {
	return &AccountService{
		cache:             dao.Rc,
		transaction:       dao.NewTransaction(),
		accountDomain:     domain.NewAccountDomain(),
		projectAuthDomain: domain.NewProjectAuthDomain(),
	}
}

// Account 获取账单和权限列表
func (a *AccountService) Account(ctx context.Context, msg *account.AccountReqMessage) (*account.AccountResponse, error) {
	//1. 去account表查询account
	//2. 去auth表查询authList
	// 获取账单列表
	accountList, total, err := a.accountDomain.AccountList(
		msg.OrganizationCode,
		msg.MemberId,
		msg.Page,
		msg.PageSize,
		msg.DepartmentCode,
		msg.SearchType)
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	// 获取权限列表
	organizationCodeId, _ := business.StringToInt32(msg.OrganizationCode)
	authList, err := a.projectAuthDomain.AuthList(int64(organizationCodeId))
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	// 拷贝accountList
	var maList []*account.MemberAccount
	copier.Copy(&maList, accountList)
	// 拷贝权限列表
	var prList []*account.ProjectAuth
	copier.Copy(&prList, authList)
	return &account.AccountResponse{
		AccountList: maList,
		AuthList:    prList,
		Total:       total,
	}, nil
}
