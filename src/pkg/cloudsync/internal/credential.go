package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// OAuthCredential refresh_token 类凭据，支持 refresh_token|client_id|client_secret 格式
type OAuthCredential struct {
	RefreshToken string
	ClientID     string
	ClientSecret string
}

// ParseOAuthCredential 解析 OAuth 凭据字符串
func ParseOAuthCredential(raw string, defaultClientID, defaultClientSecret string) OAuthCredential {
	parts := strings.SplitN(strings.TrimSpace(raw), "|", 3)
	cred := OAuthCredential{
		RefreshToken: parts[0],
		ClientID:     defaultClientID,
		ClientSecret: defaultClientSecret,
	}
	if len(parts) >= 3 && parts[1] != "" {
		cred.ClientID = parts[1]
		cred.ClientSecret = parts[2]
	}
	return cred
}

// FormatOAuthCredential 序列化 OAuth 凭据字符串
func FormatOAuthCredential(refreshToken, clientID, clientSecret, defaultClientID, defaultClientSecret string) string {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return ""
	}
	if clientID != "" && clientSecret != "" &&
		(clientID != defaultClientID || clientSecret != defaultClientSecret) {
		return refreshToken + "|" + clientID + "|" + clientSecret
	}
	return refreshToken
}

// HashCredential 对凭据做 SHA256 摘要，用于缓存键
func HashCredential(raw string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:8])
}

// BuildCredentialFromFields 根据表单字段组装凭据字符串
func BuildCredentialFromFields(authType string, fields map[string]string) string {
	switch authType {
	case "refresh_token":
		rt := strings.TrimSpace(fields["refresh_token"])
		cid := strings.TrimSpace(fields["client_id"])
		csec := strings.TrimSpace(fields["client_secret"])
		if rt == "" {
			return ""
		}
		if cid != "" && csec != "" {
			return rt + "|" + cid + "|" + csec
		}
		return rt
	default:
		if v := strings.TrimSpace(fields["cookie"]); v != "" {
			return v
		}
		if v := strings.TrimSpace(fields["authorization"]); v != "" {
			return v
		}
		for _, v := range fields {
			if strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
		return ""
	}
}
