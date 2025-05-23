package domain

import (
	"context"
	"project-grpc/user/login"
	"project-project/internal/rpc"
	"time"
)

type UserRpcDomain struct {
	lc login.LoginServiceClient
}

func NewUserRpcDomain() *UserRpcDomain {
	return &UserRpcDomain{
		lc: rpc.LoginServiceClient,
	}
}

// MemberList 根据成员id列表获取成员信息
func (d *UserRpcDomain) MemberList(mIdList []int64) ([]*login.MemberMessage, map[int64]*login.MemberMessage, error) {
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	messageList, err := d.lc.FindMemInfoByIds(c, &login.UserMessage{MIds: mIdList})
	mMap := make(map[int64]*login.MemberMessage)
	for _, v := range messageList.List {
		mMap[v.Id] = v
	}
	return messageList.List, mMap, err
}

// MemberInfo 根据成员id获取成员信息
func (d *UserRpcDomain) MemberInfo(ctx context.Context, memberCode int64) (*login.MemberMessage, error) {
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	memberMessage, err := d.lc.FindMemInfoById(c, &login.UserMessage{MemId: memberCode})
	return memberMessage, err
}
