package auth_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"project-common/encrypts"
	"project-common/errs"
	"project-grpc/auth"
	"project-project/internal/dao"
	"project-project/internal/database/tran"
	"project-project/internal/domain"
	"project-project/internal/repo"
)

type AuthService struct {
	auth.UnimplementedAuthServiceServer
	cache             repo.Cache
	transaction       tran.Transaction
	projectAuthDomain *domain.ProjectAuthDomain
}

func New() *AuthService {
	return &AuthService{
		cache:             dao.Rc,
		transaction:       dao.NewTransaction(),
		projectAuthDomain: domain.NewProjectAuthDomain(),
	}
}

// AuthList 获取权限列表
func (a *AuthService) AuthList(ctx context.Context, msg *auth.AuthReqMessage) (*auth.ListAuthMessage, error) {
	organizationCode := encrypts.DecryptNoErr(msg.OrganizationCode)
	// 获取权限列表
	listPage, total, err := a.projectAuthDomain.AuthListPage(organizationCode, msg.Page, msg.PageSize)
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	var prList []*auth.ProjectAuth
	copier.Copy(&prList, listPage)
	return &auth.ListAuthMessage{List: prList, Total: total}, nil
}
