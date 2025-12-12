package auth

import (
	"fmt"
	"myobj/src/config"
	"myobj/src/core/domain/response"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims 自定义声明
type CustomClaims struct {
	UserID    string                     `json:"user_id"`
	SessionID string                     `json:"session_id"` // 会话ID
	UserLogin response.UserLoginResponse `json:"user_login"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成 JWT
func GenerateJWT(userID string, sessionID string, userLogin response.UserLoginResponse) (string, error) {
	claims := CustomClaims{
		userID,
		sessionID,
		userLogin,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.CONFIG.Auth.JwtExpire) * time.Hour)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                                              // 签发时间
			Issuer:    "wind",                                                                                      // 签发者
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.CONFIG.Auth.Secret))
}

// ParseToken 验证 JWT
func ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.CONFIG.Auth.Secret), nil
	})
	if token == nil {
		return nil, fmt.Errorf("未授权:Token解析失败 - %v", err)
	}
	return transition(token), err
}

func transition(token *jwt.Token) *CustomClaims {
	return token.Claims.(*CustomClaims)
}
