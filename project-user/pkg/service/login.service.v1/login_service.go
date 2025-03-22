package login_service_v1

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	common "project-common"
	"project-common/encrypts"
	"project-common/errs"
	"project-common/jwts"
	"project-common/tms"
	"project-grpc/user/login"
	"project-user/config"
	"project-user/internal/dao"
	"project-user/internal/data/member"
	"project-user/internal/data/organization"
	"project-user/internal/database"
	"project-user/internal/database/tran"
	"project-user/internal/repo"
	"project-user/pkg/model"
	"strconv"
	"strings"
	"time"
)

// LoginService 提供了登录服务的实现，继承了 login.UnimplementedLoginServiceServer 的方法。
// 它通过集成缓存、成员仓库、组织仓库和事务处理来实现登录相关的功能。
type LoginService struct {
	login.UnimplementedLoginServiceServer                       // 继承自登录服务的未实现方法，为登录服务提供默认实现。
	cache                                 repo.Cache            // 缓存接口，用于快速存储和检索数据。
	memberRepo                            repo.MemberRepo       // 成员仓库接口，用于处理与成员相关的数据操作。
	organizationRepo                      repo.OrganizationRepo // 组织仓库接口，用于处理与组织相关的数据操作。
	transaction                           tran.Transaction      // 事务处理接口，用于处理需要事务支持的操作。
}

// New 创建并返回一个新的 LoginService 实例。
// 它初始化了 LoginService 结构体，并为其各个字段提供了实际的实现。
func New() *LoginService {
	// 返回一个新的 LoginService 实例，并为各个字段赋值。
	// dao.Rc 提供了缓存的实现，而 NewMemberDao、NewOrganizationDao 和 NewTransaction 分别提供了成员、组织和事务处理的实际实现。
	return &LoginService{
		cache:            dao.Rc,
		memberRepo:       dao.NewMemberDao(),
		organizationRepo: dao.NewOrganizationDao(),
		transaction:      dao.NewTransaction(),
	}
}

// GetCaptcha 处理获取验证码的请求。
func (ls *LoginService) GetCaptcha(ctx context.Context, msg *login.CaptchaMessage) (*login.CaptchaResponse, error) {
	//1.获取参数
	mobile := msg.Mobile
	//2.校验参数
	if !common.VerifyMobile(mobile) {
		return nil, errs.GrpcError(model.NoLegalMobile)
	}
	//3.生成验证码（随机4位1000-9999或者6位100000-999999）
	code := "123456"
	//4.调用短信平台（三方 放入go协程中执行 接口可以快速响应）
	go func() {
		time.Sleep(2 * time.Second)
		zap.L().Info("短信平台调用成功，发送短信")
		//redis 假设后续缓存可能存在mysql当中，也可能存在mongo当中 也可能存在memcache当中
		//5.存储验证码 redis当中 过期时间15分钟
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		// 将对应的验证码存储到 Redis 中，并设置过期时间为 15 分钟。
		err := ls.cache.Put(c, model.RegisterRedisKey+mobile, code, 15*time.Minute)
		if err != nil {
			zap.L().Info(fmt.Sprintf("验证码存入redis出错,cause by: %v \n", err))
		}
	}()
	// 返回验证码
	return &login.CaptchaResponse{Code: code}, nil
}

// Register 用户注册函数
// 该函数处理用户注册请求，包括验证参数、验证码校验、检查用户信息是否已存在，
// 以及将用户信息保存到数据库中，并创建对应的个人组织。
func (ls *LoginService) Register(ctx context.Context, msg *login.RegisterMessage) (*login.RegisterResponse, error) {
	// 初始化一个新的上下文对象，用于后续的数据库和缓存操作
	c := context.Background()

	// 1. 可以校验参数
	// 2. 校验验证码
	// 从缓存中获取验证码，检查是否存在以及是否匹配用户提交的验证码（是否过期）
	redisCode, err := ls.cache.Get(c, model.RegisterRedisKey+msg.Mobile)
	if err == redis.Nil {
		return nil, errs.GrpcError(model.CaptchaNotExist)
	}
	if err != nil {
		zap.L().Error("Register redis get error", zap.Error(err))
		return nil, errs.GrpcError(model.RedisError)
	}
	// 比较验证码
	if redisCode != msg.Captcha {
		return nil, errs.GrpcError(model.CaptchaError)
	}

	// 3. 校验业务逻辑（邮箱是否被注册 账号是否被注册 手机号是否被注册）
	// 检查邮箱、账号和手机号是否已经存在于数据库中
	exist, err := ls.memberRepo.GetMemberByEmail(c, msg.Email)
	if err != nil {
		zap.L().Error("Register db get error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if exist {
		return nil, errs.GrpcError(model.EmailExist)
	}
	// 检查账号是否已经存在于数据库中
	exist, err = ls.memberRepo.GetMemberByAccount(c, msg.Name)
	if err != nil {
		zap.L().Error("Register db get error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if exist {
		return nil, errs.GrpcError(model.AccountExist)
	}
	// 检查手机号是否已经存在于数据库中
	exist, err = ls.memberRepo.GetMemberByMobile(c, msg.Mobile)
	if err != nil {
		zap.L().Error("Register db get error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if exist {
		return nil, errs.GrpcError(model.MobileExist)
	}

	// 4. 执行业务 将数据存入member表 生成一个数据 存入组织表 organization
	// 对用户信息进行加密处理，并保存到数据库中。
	pwd := encrypts.Md5(msg.Password)
	// 创建一个新的成员对象，并设置相关属性。
	mem := &member.Member{
		Account:       msg.Name,
		Password:      pwd,
		Name:          msg.Name,
		Mobile:        msg.Mobile,
		Email:         msg.Email,
		CreateTime:    time.Now().UnixMilli(),
		LastLoginTime: time.Now().UnixMilli(),
		Status:        model.Normal,
	}
	// 在事务中执行数据库操作，包括将用户信息存入member表和创建组织。
	err = ls.transaction.Action(func(conn database.DbConn) error {
		// 存入member
		err = ls.memberRepo.SaveMember(conn, c, mem)
		if err != nil {
			zap.L().Error("Register db SaveMember error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}

		// 存入组织
		// 创建用户的个人组织，并保存到数据库
		org := &organization.Organization{
			Name:       mem.Name + "个人组织",
			MemberId:   mem.Id,
			CreateTime: time.Now().UnixMilli(),
			Personal:   model.Personal,
			Avatar:     "https://gimg2.baidu.com/image_search/src=http%3A%2F%2Fc-ssl.dtstatic.com%2Fuploads%2Fblog%2F202103%2F31%2F20210331160001_9a852.thumb.1000_0.jpg&refer=http%3A%2F%2Fc-ssl.dtstatic.com&app=2002&size=f9999,10000&q=a80&n=0&g=0n&fmt=auto?sec=1673017724&t=ced22fc74624e6940fd6a89a21d30cc5",
		}
		// 调用 OrganizationRepo 的 SaveOrganization 方法将组织信息存入数据库。
		err = ls.organizationRepo.SaveOrganization(conn, c, org)
		if err != nil {
			zap.L().Error("register SaveOrganization db err", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	})

	// 5. 返回
	// 返回注册响应，如果操作成功，返回空错误；否则，返回遇到的错误
	return &login.RegisterResponse{}, err
}

// Login 实现登录服务
// 该方法接收登录信息，验证用户身份，生成并返回登录响应，包括用户信息、组织信息和令牌信息
func (ls *LoginService) Login(ctx context.Context, msg *login.LoginMessage) (*login.LoginResponse, error) {
	// 创建一个新的上下文对象，用于后续的数据库查询等操作
	c := context.Background()

	// 1. 去数据库查询 账号密码是否正确
	// 对输入的密码进行MD5加密，以便与数据库中的密码进行比较
	pwd := encrypts.Md5(msg.Password)
	// 调用成员仓库的FindMember方法查询数据库中是否存在匹配的用户名和密码
	mem, err := ls.memberRepo.FindMember(c, msg.Account, pwd)
	if err != nil {
		// 如果查询过程中出现错误，记录错误日志并返回数据库错误
		zap.L().Error("Login db FindMember error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if mem == nil {
		// 如果查询结果为空，说明用户名或密码不正确，返回相应错误
		return nil, errs.GrpcError(model.AccountAndPwdError)
	}

	// 将查询到的成员信息复制到MemberMessage对象中，并对成员ID进行加密
	memMsg := &login.MemberMessage{}
	err = copier.Copy(memMsg, mem)
	memMsg.Code, _ = encrypts.EncryptInt64(mem.Id, model.AESKey)

	// 2. 根据用户id查组织
	// 调用组织仓库的FindOrganizationByMemId方法，根据成员ID查询其所属的组织信息
	orgs, err := ls.organizationRepo.FindOrganizationByMemId(c, mem.Id)
	if err != nil {
		// 如果查询过程中出现错误，记录错误日志并返回数据库错误
		zap.L().Error("Login db FindMember error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}

	// 将查询到的组织信息复制到OrganizationsMessage列表中，并对每个组织的ID进行加密
	var orgsMessage []*login.OrganizationMessage
	err = copier.Copy(&orgsMessage, orgs)
	for _, v := range orgsMessage {
		v.Code, _ = encrypts.EncryptInt64(v.Id, model.AESKey)
		v.OwnerCode = memMsg.Code
		o := organization.ToMap(orgs)[v.Id]
		v.CreateTime = tms.FormatByMill(o.CreateTime)
	}
	if len(orgs) > 0 {
		memMsg.OrganizationCode, _ = encrypts.EncryptInt64(orgs[0].Id, model.AESKey)
	}
	// 3. 用jwt生成token
	// 将成员ID转换为字符串，用于生成JWT令牌
	memIdStr := strconv.FormatInt(mem.Id, 10)
	// 计算访问令牌和刷新令牌的过期时间
	exp := time.Duration(config.C.JwtConfig.AccessExp*3600*24) * time.Second
	rExp := time.Duration(config.C.JwtConfig.RefreshExp*3600*24) * time.Second
	// 生成JWT令牌
	token := jwts.CreateToken(memIdStr, exp, config.C.JwtConfig.AccessSecret, rExp, config.C.JwtConfig.RefreshSecret)
	// 将生成的令牌信息封装到TokenMessage对象中
	tokenList := &login.TokenMessage{
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		AccessTokenExp: token.AccessExp,
		TokenType:      "bearer",
	}

	// 返回登录响应，包括成员信息、组织信息和令牌信息
	return &login.LoginResponse{
		Member:           memMsg,
		OrganizationList: orgsMessage,
		TokenList:        tokenList,
	}, nil
}

// TokenVerify 验证用户登录状态
// 该方法接收一个LoginMessage，其中包含用户提供的token信息
// 它会解析token，验证其有效性，并从数据库中获取用户信息
// 如果验证成功，返回包含用户信息的LoginResponse；如果失败，返回错误
func (ls *LoginService) TokenVerify(ctx context.Context, msg *login.LoginMessage) (*login.LoginResponse, error) {
	// 提取token信息，并处理带有bearer前缀的token
	token := msg.Token
	if strings.Contains(token, "bearer") {
		token = strings.ReplaceAll(token, "bearer ", "")
	}

	// 解析token，验证其有效性
	parseToken, err := jwts.ParseToken(token, config.C.JwtConfig.AccessSecret)
	if err != nil {
		// 如果token验证失败，记录错误日志，并返回登录错误
		zap.L().Error("Login  TokenVerify error", zap.Error(err))
		return nil, errs.GrpcError(model.NoLogin)
	}

	// 将解析后的token转换为用户ID
	id, _ := strconv.ParseInt(parseToken, 10, 64)

	// 根据用户ID从数据库中查询用户信息
	// 注意：这里可以进行优化，例如在用户登录后缓存用户信息，以减少数据库查询
	memberById, err := ls.memberRepo.FindMemberById(context.Background(), id)
	if err != nil {
		// 如果数据库查询失败，记录错误日志，并返回数据库错误
		zap.L().Error("TokenVerify db FindMemberById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}

	// 将查询到的用户信息复制到新的MemberMessage对象中，并加密用户ID
	memMsg := &login.MemberMessage{}
	copier.Copy(memMsg, memberById)
	// 加密用户ID
	memMsg.Code, _ = encrypts.EncryptInt64(memberById.Id, model.AESKey)

	// 返回包含用户信息的登录响应
	return &login.LoginResponse{Member: memMsg}, nil
}

// MyOrgList 获取用户所属的组织列表
// 该方法根据用户ID查询数据库中相关的组织信息，并返回组织列表
// 参数:
func (l *LoginService) MyOrgList(ctx context.Context, msg *login.UserMessage) (*login.OrgListResponse, error) {
	// 提取用户ID
	memId := msg.MemId

	// 调用组织仓库的FindOrganizationByMemId方法查询用户所属的组织
	orgs, err := l.organizationRepo.FindOrganizationByMemId(ctx, memId)
	if err != nil {
		// 如果查询过程中出现错误，记录错误日志并返回相应的gRPC错误
		zap.L().Error("MyOrgList FindOrganizationByMemId err", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}

	// 初始化一个组织消息列表，用于存储转换后的组织信息
	var orgsMessage []*login.OrganizationMessage

	// 将查询到的组织信息复制到组织消息列表中
	err = copier.Copy(&orgsMessage, orgs)

	// 对每个组织的消息进行处理，加密组织ID
	for _, org := range orgsMessage {
		// 使用AES密钥加密组织ID，并将加密后的结果赋值给组织的Code字段
		org.Code, _ = encrypts.EncryptInt64(org.Id, model.AESKey)
	}

	// 构建并返回包含组织消息列表的响应对象
	return &login.OrgListResponse{OrganizationList: orgsMessage}, nil
}

func (ls *LoginService) FindMemInfoById(ctx context.Context, msg *login.UserMessage) (*login.MemberMessage, error) {
	memberById, err := ls.memberRepo.FindMemberById(context.Background(), msg.MemId)
	if err != nil {
		zap.L().Error("TokenVerify db FindMemberById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	memMsg := &login.MemberMessage{}
	copier.Copy(memMsg, memberById)
	memMsg.Code, _ = encrypts.EncryptInt64(memberById.Id, model.AESKey)
	orgs, err := ls.organizationRepo.FindOrganizationByMemId(context.Background(), memberById.Id)
	if err != nil {
		zap.L().Error("TokenVerify db FindMember error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if len(orgs) > 0 {
		memMsg.OrganizationCode, _ = encrypts.EncryptInt64(orgs[0].Id, model.AESKey)
	}
	memMsg.CreateTime = tms.FormatByMill(memberById.CreateTime)
	return memMsg, nil
}
