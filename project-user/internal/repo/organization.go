package repo

import (
	"context"
	"project-user/internal/data/organization"
	"project-user/internal/database"
)

type OrganizationRepo interface {
	// SaveOrganization 保存组织
	SaveOrganization(conn database.DbConn, ctx context.Context, org *organization.Organization) error
	// FindOrganizationByMemId 根据成员id查询组织
	FindOrganizationByMemId(ctx context.Context, memId int64) ([]*organization.Organization, error)
}
