package repo

import (
	"context"
	"project-user/internal/data/member"
	"project-user/internal/database"
)

type MemberRepo interface {
	// GetMemberByEmail 检查邮箱、账号和手机号是否已经存在于数据库中
	GetMemberByEmail(ctx context.Context, email string) (bool, error)
	// GetMemberByAccount 检查账号是否已经存在于数据库中
	GetMemberByAccount(ctx context.Context, account string) (bool, error)
	// GetMemberByMobile 检查手机号是否已经存在于数据库中
	GetMemberByMobile(ctx context.Context, mobile string) (bool, error)
	// SaveMember 保存会员信息
	SaveMember(conn database.DbConn, ctx context.Context, mem *member.Member) error
	// FindMember 根据账号和密码查找会员信息
	FindMember(ctx context.Context, account string, pwd string) (mem *member.Member, err error)
	// FindMemberById 根据会员ID查找会员信息
	FindMemberById(background context.Context, id int64) (mem *member.Member, err error)
	// FindMemberByIds 根据会员ID列表查找会员信息
	FindMemberByIds(background context.Context, ids []int64) (list []*member.Member, err error)
}
