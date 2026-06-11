package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"myobj/src/config"
	"myobj/src/core/domain/response"
	"myobj/src/pkg/logger"
	"net"
	"strings"
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

// getSecret 获取 JWT 密钥，校验长度至少 32 字节，不足时自动生成随机密钥
func getSecret() (string, error) {
	secret := config.CONFIG.Auth.Secret
	if len(secret) < 32 {
		logger.LOG.Warn("JWT Secret 长度不足32字节，将自动生成随机密钥（注意：重启后旧 token 将失效）")
		secretBytes := make([]byte, 32)
		if _, err := rand.Read(secretBytes); err != nil {
			return "", fmt.Errorf("failed to generate JWT secret: %w", err)
		}
		return hex.EncodeToString(secretBytes), nil
	}
	return secret, nil
}

// GenerateJWT 生成 JWT
func GenerateJWT(userID string, sessionID string, userLogin response.UserLoginResponse) (string, error) {
	jwtExpire := config.CONFIG.Auth.JwtExpire
	if jwtExpire <= 0 || jwtExpire > 720 {
		jwtExpire = 720
	}
	claims := CustomClaims{
		userID,
		sessionID,
		userLogin,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwtExpire) * time.Hour)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                           // 签发时间
			Issuer:    "wind",                                                                   // 签发者
		},
	}
	secret, err := getSecret()
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 验证 JWT
func ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		secret, err := getSecret()
		if err != nil {
			return nil, err
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("未授权:Token解析失败 - %v", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}

// GetCookieDomain 获取 Cookie 的 domain
func GetCookieDomain(host string) string {
	// 如果是本地开发环境（精确匹配，避免误匹配包含这些子串的域名）
	if host == "localhost" || strings.HasPrefix(host, "localhost:") {
		return "localhost"
	}

	// 提取主机名（去除端口）
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}

	// IP 地址不设置 Domain 属性，防止子域 Cookie 注入风险
	if net.ParseIP(hostname) != nil {
		return ""
	}

	// 解析域名，设置上级域
	// 例如：api.example.com → .example.com
	parts := strings.Split(hostname, ".")
	if len(parts) >= 2 {
		return "." + strings.Join(parts[len(parts)-2:], ".")
	}

	return ""
}
