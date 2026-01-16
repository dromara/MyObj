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
	"sync"
	"time"
)

// SignatureV4 AWS Signature V4验证器
type SignatureV4 struct {
	region  string
	service string
	// 密钥缓存，提高性能
	keyCache *signingKeyCache
}

// signingKeyCache 签名密钥缓存
type signingKeyCache struct {
	mu    sync.RWMutex
	cache map[string]cachedKey
}

// cachedKey 缓存的密钥
type cachedKey struct {
	key      []byte
	date     string
	expires  time.Time
}

// newSigningKeyCache 创建新的密钥缓存
func newSigningKeyCache() *signingKeyCache {
	return &signingKeyCache{
		cache: make(map[string]cachedKey),
	}
}

// get 从缓存获取密钥
func (c *signingKeyCache) get(key string, dateStamp string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	entry, ok := c.cache[key]
	if !ok {
		return nil, false
	}
	
	// 检查日期是否匹配（同一天）
	if entry.date != dateStamp {
		return nil, false
	}
	
	// 检查是否过期（缓存有效期24小时）
	if time.Now().After(entry.expires) {
		return nil, false
	}
	
	return entry.key, true
}

// set 设置缓存
func (c *signingKeyCache) set(key string, dateStamp string, signingKey []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 缓存24小时
	c.cache[key] = cachedKey{
		key:     signingKey,
		date:    dateStamp,
		expires: time.Now().Add(24 * time.Hour),
	}
	
	// 清理过期缓存（简单策略：如果缓存超过100个，清理一半）
	if len(c.cache) > 100 {
		now := time.Now()
		for k, v := range c.cache {
			if now.After(v.expires) {
				delete(c.cache, k)
			}
		}
	}
}

// NewSignatureV4 创建签名验证器
func NewSignatureV4(region, service string) *SignatureV4 {
	return &SignatureV4{
		region:   region,
		service:  service,
		keyCache: newSigningKeyCache(),
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

	queryString := strings.Join(parts, "&")
	// AWS 规范要求：将 + 替换为 %20（空格）
	// url.QueryEscape 会将空格编码为 +，但 AWS 要求使用 %20
	queryString = strings.Replace(queryString, "+", "%20", -1)
	
	return queryString
}

// stripExcessSpaces 移除多余空格（参考 AWS SDK v2 实现）
func stripExcessSpaces(str string) string {
	// 移除尾随空格
	j := len(str) - 1
	for j >= 0 && str[j] == ' ' {
		j--
	}
	
	// 移除前导空格
	k := 0
	for k < j && str[k] == ' ' {
		k++
	}
	
	if k > 0 || j < len(str)-1 {
		str = str[k : j+1]
	}
	
	// 移除多个连续空格
	if strings.Contains(str, "  ") {
		buf := []byte(str)
		var result []byte
		var lastSpace bool
		
		for _, b := range buf {
			if b == ' ' {
				if !lastSpace {
					result = append(result, b)
					lastSpace = true
				}
			} else {
				result = append(result, b)
				lastSpace = false
			}
		}
		str = string(result)
	}
	
	return str
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
		
		// 规范化：使用更严格的空格处理（参考 AWS SDK v2）
		value = strings.TrimSpace(stripExcessSpaces(value))
		
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

// deriveSigningKey 派生签名密钥（带缓存）
func (sv *SignatureV4) deriveSigningKey(secretKey, dateStamp string) []byte {
	// 构建缓存键：secretKey + region + service + dateStamp
	cacheKey := fmt.Sprintf("%s:%s:%s:%s", secretKey, sv.region, sv.service, dateStamp)
	
	// 尝试从缓存获取
	if cachedKey, ok := sv.keyCache.get(cacheKey, dateStamp); ok {
		return cachedKey
	}
	
	// 计算签名密钥
	kDate := sv.hmacSHA256([]byte("AWS4"+secretKey), []byte(dateStamp))
	kRegion := sv.hmacSHA256(kDate, []byte(sv.region))
	kService := sv.hmacSHA256(kRegion, []byte(sv.service))
	kSigning := sv.hmacSHA256(kService, []byte("aws4_request"))
	
	// 存入缓存
	sv.keyCache.set(cacheKey, dateStamp, kSigning)
	
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
	
	// 构建规范 headers（支持更多 headers，如 X-Amz-Acl）
	canonicalHeaders := fmt.Sprintf("host:%s\n", strings.ToLower(parsedURL.Host))
	signedHeadersList := []string{"host"}
	
	// 处理额外的 headers（Header Hoisting）
	if params.Headers != nil {
		for k, v := range params.Headers {
			lowerKey := strings.ToLower(k)
			// 支持 X-Amz-* headers
			if strings.HasPrefix(lowerKey, "x-amz-") {
				// 将 header 添加到查询参数（Header Hoisting）
				query.Set(k, v)
				// 添加到规范 headers
				canonicalHeaders += fmt.Sprintf("%s:%s\n", lowerKey, strings.TrimSpace(v))
				signedHeadersList = append(signedHeadersList, lowerKey)
			}
		}
	}
	
	sort.Strings(signedHeadersList)
	signedHeaders := strings.Join(signedHeadersList, ";")
	query.Set("X-Amz-SignedHeaders", signedHeaders)

	// 6. 构建规范请求
	canonicalURI := parsedURL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	canonicalQueryString := sv.buildCanonicalQueryString(query)
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

	// 检查是否过期（修复：使用正确的过期时间计算）
	// requestTime 是签名时间，expiresSeconds 是相对过期时间（秒）
	expiresAt := requestTime.Add(time.Duration(expiresSeconds) * time.Second)
	if time.Now().After(expiresAt) {
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
