package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
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
		// 返回更详细的错误信息用于调试
		return fmt.Errorf("signature mismatch: calculated=%s, provided=%s", signature, providedSignature)
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
		headerName := strings.ToLower(strings.TrimSpace(h))
		
		// 获取 header 值（HTTP header 名称是大小写不敏感的）
		value := ""
		// 先尝试使用原始名称
		value = r.Header.Get(h)
		// 如果为空，尝试使用小写名称
		if value == "" {
			value = r.Header.Get(headerName)
		}
		
		// 特殊处理 host 头：确保格式正确
		if headerName == "host" {
			// 如果请求中没有 host 头，使用 r.Host
			if value == "" {
				value = r.Host
			}
			// 规范化 host：转小写
			value = strings.ToLower(value)
		}
		
		// 规范化：移除多余空格，合并连续空格为单个空格
		value = strings.TrimSpace(value)
		value = strings.Join(strings.Fields(value), " ")
		
		parts = append(parts, headerName+":"+value)
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

// PresignedURLParams 预签名URL参数
type PresignedURLParams struct {
	Method      string            // HTTP方法（GET, PUT等）
	Bucket      string            // Bucket名称
	Key         string            // Object Key
	Expires     int64             // 过期时间（秒）
	AccessKeyID string            // Access Key ID
	SecretKey   string            // Secret Key
	Headers     map[string]string // 额外的请求头
}

// GeneratePresignedURL 生成预签名URL
func (sv *SignatureV4) GeneratePresignedURL(baseURL string, params PresignedURLParams) (string, error) {
	// 1. 构建URL
	objectURL := fmt.Sprintf("%s/%s/%s", baseURL, params.Bucket, params.Key)
	parsedURL, err := url.Parse(objectURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// 2. 计算过期时间
	expiresAt := time.Now().Add(time.Duration(params.Expires) * time.Second)
	expiresSeconds := strconv.FormatInt(params.Expires, 10)

	// 3. 构建日期时间戳
	dateTime := expiresAt.UTC().Format("20060102T150405Z")
	dateStamp := dateTime[:8]

	// 4. 构建凭证范围
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, sv.region, sv.service)
	credential := fmt.Sprintf("%s/%s", params.AccessKeyID, credentialScope)

	// 5. 构建查询参数
	query := parsedURL.Query()
	query.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	query.Set("X-Amz-Credential", credential)
	query.Set("X-Amz-Date", dateTime)
	query.Set("X-Amz-Expires", expiresSeconds)
	query.Set("X-Amz-SignedHeaders", "host")

	// 6. 构建规范请求
	canonicalURI := parsedURL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	canonicalQueryString := sv.buildCanonicalQueryString(query)
	canonicalHeaders := fmt.Sprintf("host:%s\n", parsedURL.Host)
	signedHeaders := "host"
	payloadHash := "UNSIGNED-PAYLOAD"

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		params.Method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	)

	// 7. 构建待签名字符串
	stringToSign := sv.buildStringToSign(dateTime, credentialScope, canonicalRequest)

	// 8. 计算签名
	signingKey := sv.deriveSigningKey(params.SecretKey, dateStamp)
	signature := sv.calculateSignature(signingKey, stringToSign)

	// 9. 添加签名到查询参数
	query.Set("X-Amz-Signature", signature)
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

// VerifyPresignedURL 验证预签名URL
func (sv *SignatureV4) VerifyPresignedURL(r *http.Request, secretKey string) error {
	// 1. 提取查询参数
	query := r.URL.Query()
	algorithm := query.Get("X-Amz-Algorithm")
	credential := query.Get("X-Amz-Credential")
	dateTime := query.Get("X-Amz-Date")
	expires := query.Get("X-Amz-Expires")
	signedHeaders := query.Get("X-Amz-SignedHeaders")
	signature := query.Get("X-Amz-Signature")

	// 2. 验证必需参数
	if algorithm != "AWS4-HMAC-SHA256" {
		return fmt.Errorf("unsupported algorithm")
	}
	if credential == "" || dateTime == "" || expires == "" || signedHeaders == "" || signature == "" {
		return fmt.Errorf("missing required query parameters")
	}

	// 3. 验证过期时间
	expiresSeconds, err := strconv.ParseInt(expires, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid expires parameter")
	}

	requestTime, err := time.Parse("20060102T150405Z", dateTime)
	if err != nil {
		return fmt.Errorf("invalid date format")
	}

	// 检查是否过期
	if time.Since(requestTime) > time.Duration(expiresSeconds)*time.Second {
		return fmt.Errorf("presigned URL expired")
	}

	// 4. 提取日期戳
	if len(dateTime) < 8 {
		return fmt.Errorf("invalid date format")
	}
	dateStamp := dateTime[:8]

	// 5. 构建规范请求
	canonicalURI := r.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	// 构建查询字符串（排除签名）
	canonicalQuery := make(url.Values)
	for k, v := range query {
		if k != "X-Amz-Signature" {
			canonicalQuery[k] = v
		}
	}
	canonicalQueryString := sv.buildCanonicalQueryString(canonicalQuery)

	canonicalHeaders := fmt.Sprintf("host:%s\n", r.Host)
	payloadHash := "UNSIGNED-PAYLOAD"

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		r.Method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	)

	// 6. 构建待签名字符串
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, sv.region, sv.service)
	stringToSign := sv.buildStringToSign(dateTime, credentialScope, canonicalRequest)

	// 7. 计算签名
	signingKey := sv.deriveSigningKey(secretKey, dateStamp)
	calculatedSignature := sv.calculateSignature(signingKey, stringToSign)

	// 8. 比较签名
	if calculatedSignature != signature {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}
