package jwts

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// JwtToken 结构体用于存储访问令牌和刷新令牌及其过期时间。
type JwtToken struct {
	AccessToken  string // 访问令牌
	RefreshToken string // 刷新令牌
	AccessExp    int64  // 访问令牌过期时间
	RefreshExp   int64  // 刷新令牌过期时间
}

// CreateToken 生成访问令牌和刷新令牌。
// 参数:
//
//	val: 令牌的值
//	exp: 访问令牌的过期时间
//	secret: 访问令牌的密钥
//	refreshExp: 刷新令牌的过期时间
//	refreshSecret: 刷新令牌的密钥
//
// 返回值:
//
//	*JwtToken: 生成的令牌结构体指针
func CreateToken(val string, exp time.Duration, secret string, refreshExp time.Duration, refreshSecret string, ip string) *JwtToken {
	// 计算访问令牌的过期时间
	aExp := time.Now().Add(exp).Unix()
	// 创建访问令牌
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token": val,
		"exp":   aExp,
		"ip":    ip,
	})
	// 签发访问令牌
	aToken, _ := accessToken.SignedString([]byte(secret))

	// 计算刷新令牌的过期时间
	rExp := time.Now().Add(refreshExp).Unix()
	// 创建刷新令牌
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token": val,
		"exp":   aExp,
	})
	// 签发刷新令牌
	rToken, _ := refreshToken.SignedString([]byte(refreshSecret))

	// 返回令牌结构体
	return &JwtToken{
		AccessExp:    aExp,
		AccessToken:  aToken,
		RefreshExp:   rExp,
		RefreshToken: rToken,
	}
}

func ParseTokenOld(tokenString string, secret string) (string, error) {
	// 解析令牌
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证令牌的签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// 返回令牌的密钥
		return []byte(secret), nil
	})
	// 如果解析出错，返回错误信息
	if err != nil {
		return "", err
	}

	// 验证令牌的有效性
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 获取令牌的值和过期时间
		val := claims["token"].(string)
		exp := int64(claims["exp"].(float64))
		// 如果令牌过期，返回错误信息
		if exp <= time.Now().Unix() {
			return "", errors.New("token过期了")
		}
		// 返回令牌的值
		return val, nil
	} else {
		// 如果令牌无效，返回错误信息
		return "", err
	}
}

// ParseToken 解析令牌并验证其有效性。
// 参数:
//
//	tokenString: 令牌字符串
//	secret: 令牌的密钥
//
// 返回值:
//
//	string: 令牌的值
//	error: 错误信息，如果有的话
func ParseToken(tokenString string, secret string, ip string) (string, error) {
	// 解析令牌
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证令牌的签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// 返回令牌的密钥
		return []byte(secret), nil
	})
	// 如果解析出错，返回错误信息
	if err != nil {
		return "", err
	}

	// 验证令牌的有效性
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 获取令牌的值和过期时间
		val := claims["token"].(string)
		exp := int64(claims["exp"].(float64))
		// 如果令牌过期，返回错误信息
		if exp <= time.Now().Unix() {
			return "", errors.New("token过期了")
		}
		if claims["ip"] != ip {
			return "", errors.New("ip不合法")
		}
		// 返回令牌的值
		return val, nil
	} else {
		// 如果令牌无效，返回错误信息
		return "", err
	}
}
