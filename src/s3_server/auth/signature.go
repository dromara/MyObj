package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// SignatureV4 AWS Signature V4验证器
type SignatureV4 struct {
	region  string
	service string
}

// NewSignatureV4 创建签名验证器
func NewSignatureV4(region, service string) *SignatureV4 {
	return &SignatureV4{
		region:  region,
		service: service,
	}
}

// VerifyRequest 验证请求签名
func (sv *SignatureV4) VerifyRequest(r *http.Request, secretKey string) error {
	// 1. 解析Authorization头
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return fmt.Errorf("missing Authorization header")
	}

	// 检查签名版本
	if !strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256") {
		return fmt.Errorf("unsupported signature version")
	}

	// 2. 提取签名信息
	parts := sv.parseAuthHeader(authHeader)
	credential := parts["Credential"]
	signedHeaders := parts["SignedHeaders"]
	providedSignature := parts["Signature"]

	if credential == "" || signedHeaders == "" || providedSignature == "" {
		return fmt.Errorf("invalid Authorization header format")
	}

	// 3. 构建规范请求
	canonicalRequest := sv.buildCanonicalRequest(r, signedHeaders)

	// 4. 构建待签名字符串
	dateStamp := r.Header.Get("X-Amz-Date")
	if dateStamp == "" {
		dateStamp = r.Header.Get("Date")
	}
	if len(dateStamp) < 8 {
		return fmt.Errorf("invalid date format")
	}

	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request",
		dateStamp[:8], sv.region, sv.service)
	stringToSign := sv.buildStringToSign(dateStamp, credentialScope, canonicalRequest)

	// 5. 计算签名密钥
	signingKey := sv.deriveSigningKey(secretKey, dateStamp[:8])

	// 6. 计算签名
	signature := sv.calculateSignature(signingKey, stringToSign)

	// 7. 比较签名
	if signature != providedSignature {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

// parseAuthHeader 解析Authorization头
func (sv *SignatureV4) parseAuthHeader(authHeader string) map[string]string {
	result := make(map[string]string)

	// 移除前缀
	authHeader = strings.TrimPrefix(authHeader, "AWS4-HMAC-SHA256 ")

	// 分割键值对
	parts := strings.Split(authHeader, ", ")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}

	return result
}

// buildCanonicalRequest 构建规范请求
func (sv *SignatureV4) buildCanonicalRequest(r *http.Request, signedHeaders string) string {
	// 1. HTTP方法
	method := r.Method

	// 2. 规范URI
	canonicalURI := r.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	// 3. 规范查询字符串
	canonicalQueryString := sv.buildCanonicalQueryString(r.URL.Query())

	// 4. 规范请求头
	canonicalHeaders := sv.buildCanonicalHeaders(r, signedHeaders)

	// 5. 已签名的请求头列表
	signedHeadersList := signedHeaders

	// 6. 请求负载哈希
	payloadHash := r.Header.Get("X-Amz-Content-Sha256")
	if payloadHash == "" {
		payloadHash = "UNSIGNED-PAYLOAD"
	}

	// 组合规范请求
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeadersList,
		payloadHash,
	)
}

// buildCanonicalQueryString 构建规范查询字符串
func (sv *SignatureV4) buildCanonicalQueryString(query url.Values) string {
	if len(query) == 0 {
		return ""
	}

	var keys []string
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		values := query[k]
		sort.Strings(values)
		for _, v := range values {
			if v == "" {
				parts = append(parts, url.QueryEscape(k)+"=")
			} else {
				parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
			}
		}
	}

	return strings.Join(parts, "&")
}

// buildCanonicalHeaders 构建规范请求头
func (sv *SignatureV4) buildCanonicalHeaders(r *http.Request, signedHeaders string) string {
	headers := strings.Split(signedHeaders, ";")
	sort.Strings(headers)

	var parts []string
	for _, h := range headers {
		value := strings.TrimSpace(r.Header.Get(h))
		// 规范化：移除多余空格
		value = strings.Join(strings.Fields(value), " ")
		parts = append(parts, h+":"+value)
	}

	return strings.Join(parts, "\n") + "\n"
}

// buildStringToSign 构建待签名字符串
func (sv *SignatureV4) buildStringToSign(requestDateTime, credentialScope, canonicalRequest string) string {
	algorithm := "AWS4-HMAC-SHA256"
	hashedCanonicalRequest := sv.sha256Hex([]byte(canonicalRequest))

	return fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm,
		requestDateTime,
		credentialScope,
		hashedCanonicalRequest,
	)
}

// deriveSigningKey 派生签名密钥
func (sv *SignatureV4) deriveSigningKey(secretKey, dateStamp string) []byte {
	kDate := sv.hmacSHA256([]byte("AWS4"+secretKey), []byte(dateStamp))
	kRegion := sv.hmacSHA256(kDate, []byte(sv.region))
	kService := sv.hmacSHA256(kRegion, []byte(sv.service))
	kSigning := sv.hmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}

// calculateSignature 计算签名
func (sv *SignatureV4) calculateSignature(signingKey []byte, stringToSign string) string {
	signature := sv.hmacSHA256(signingKey, []byte(stringToSign))
	return hex.EncodeToString(signature)
}

// hmacSHA256 HMAC-SHA256计算
func (sv *SignatureV4) hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// sha256Hex SHA256哈希（十六进制）
func (sv *SignatureV4) sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// ExtractAccessKeyID 从Authorization头提取Access Key ID
func ExtractAccessKeyID(authHeader string) string {
	if !strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256") {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "Credential=") {
			credential := strings.TrimPrefix(part, "Credential=")
			// 提取Access Key ID（第一个斜杠前的部分）
			if idx := strings.Index(credential, "/"); idx > 0 {
				return credential[:idx]
			}
		}
	}

	return ""
}
