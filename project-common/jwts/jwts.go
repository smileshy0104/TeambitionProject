package jwts

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// JwtToken 定义了JWT令牌的结构，包括访问令牌和刷新令牌。
type JwtToken struct {
	AccessToken  string // 访问令牌
	RefreshToken string // 刷新令牌
	AccessExp    int64  // 访问令牌的过期时间
	RefreshExp   int64  // 刷新令牌的过期时间
}

// CreateToken 创建一组访问令牌和刷新令牌。
// 参数:
// val: 要存储在令牌中的值。
// exp: 访问令牌的过期时间。
// secret: 用于签名访问令牌的密钥。
// refreshExp: 刷新令牌的过期时间。
// refreshSecret: 用于签名刷新令牌的密钥。
// 返回值:
// *JwtToken: 包含访问令牌和刷新令牌的结构体指针。
func CreateToken(val string, exp time.Duration, secret string, refreshExp time.Duration, refreshSecret string) *JwtToken {
	// 计算访问令牌的过期时间
	aExp := time.Now().Add(exp).Unix()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token": val,
		"exp":   aExp,
	})
	// 签名访问令牌
	aToken, _ := accessToken.SignedString([]byte(secret))

	// 计算刷新令牌的过期时间
	rExp := time.Now().Add(refreshExp).Unix()
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token": val,
		"exp":   aExp,
	})
	// 签名刷新令牌
	rToken, _ := refreshToken.SignedString([]byte(refreshSecret))

	// 返回包含令牌信息的结构体
	return &JwtToken{
		AccessExp:    aExp,
		AccessToken:  aToken,
		RefreshExp:   rExp,
		RefreshToken: rToken,
	}
}

// ParseToken 解析并验证JWT令牌。
// 参数:
// tokenString: 要解析的JWT令牌字符串。
// secret: 用于验证令牌签名的密钥。
func ParseToken(tokenString string, secret string) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法是否符合预期
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}

		// 返回用于验证的密钥
		return []byte(secret), nil
	})

	// 如果令牌有效且解析成功，则打印声明内容
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Printf("%v \n", claims)
	} else {
		// 打印解析错误
		fmt.Println(err)
	}
}
