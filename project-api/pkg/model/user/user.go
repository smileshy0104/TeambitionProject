package user

import (
	"errors"
	common "project-common"
)

// RegisterReq 注册请求结构体
type RegisterReq struct {
	Email     string `json:"email" form:"email"`
	Name      string `json:"name" form:"name"`
	Password  string `json:"password" form:"password"`
	Password2 string `json:"password2" form:"password2"`
	Mobile    string `json:"mobile" form:"mobile"`
	Captcha   string `json:"captcha" form:"captcha"`
}

// VerifyPassword 验证密码一致性
func (r RegisterReq) VerifyPassword() bool {
	return r.Password == r.Password2
}

// Verify 验证注册信息的合法性
func (r RegisterReq) Verify() error {
	if !common.VerifyEmailFormat(r.Email) {
		return errors.New("邮箱格式不正确")
	}
	if !common.VerifyMobile(r.Mobile) {
		return errors.New("手机号格式不正确")
	}
	if !r.VerifyPassword() {
		return errors.New("两次密码输入不一致")
	}
	return nil
}

// LoginReq 登录请求结构体
type LoginReq struct {
	Account  string `json:"account" form:"account"`
	Password string `json:"password" form:"password"`
}

// LoginRsp 登录响应结构体
type LoginRsp struct {
	Member           Member             `json:"member"`
	TokenList        TokenList          `json:"tokenList"`
	OrganizationList []OrganizationList `json:"organizationList"`
}

// Member 会员信息结构体
type Member struct {
	Name             string `json:"name"`
	Mobile           string `json:"mobile"`
	Status           int    `json:"status"`
	Code             string `json:"code"`
	Email            string `json:"email"`
	CreateTime       string `json:"create_time"`
	LastLoginTime    string `json:"last_login_time"`
	OrganizationCode string `json:"organization_code"`
}

// TokenList 令牌列表结构体
type TokenList struct {
	AccessToken    string `json:"accessToken"`
	RefreshToken   string `json:"refreshToken"`
	TokenType      string `json:"tokenType"`
	AccessTokenExp int64  `json:"accessTokenExp"`
}

// OrganizationList 组织列表结构体
type OrganizationList struct {
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	OwnerCode   string `json:"owner_code"`
	CreateTime  string `json:"create_time"`
	Personal    int32  `json:"personal"`
	Address     string `json:"address"`
	Province    int32  `json:"province"`
	City        int32  `json:"city"`
	Area        int32  `json:"area"`
	Code        string `json:"code"`
}
