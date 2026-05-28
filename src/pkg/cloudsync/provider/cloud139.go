package provider

import (
	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/cloudsync/internal"
	"myobj/src/pkg/enum"

	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	cloud139TokenRefreshURL = "https://aas.caiyun.feixin.10086.cn:443/tellin/authTokenRefresh.do"
	cloud139RoutePolicyURL  = "https://user-njs.yun.139.com/user/route/qryRoutePolicy"
)

func init() {
	cloudsync.Register(cloudsync.ProviderInfo{
		ID: "139", Name: "中国移动云盘", AuthType: cloudsync.AuthCookie,
		Description: "authorization 登录（Base64 编码）",
	}, enum.DownloadTaskType139.Value(), func(cookie string) cloudsync.CloudProvider {
		return NewCloud139Provider(cookie)
	})
}

// cloud139TokenRefreshResp token 刷新 XML 响应
type cloud139TokenRefreshResp struct {
	XMLName     xml.Name `xml:"root"`
	Return      string   `xml:"return"`
	Token       string   `xml:"token"`
	AccessToken string   `xml:"accessToken"`
	Desc        string   `xml:"desc"`
}

// cloud139RoutePolicyResp 路由策略响应
type cloud139RoutePolicyResp struct {
	Code string `json:"code"`
	Data struct {
		RoutePolicyList []struct {
			ModName  string `json:"modName"`
			HttpsURL string `json:"httpsUrl"`
		} `json:"routePolicyList"`
	} `json:"data"`
}

// cloud139File 139 云盘文件
type cloud139File struct {
	FileID    string `json:"fileId"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Type      string `json:"type"` // "folder" or "file"
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// cloud139ListResp 文件列表响应
type cloud139ListResp struct {
	Code string `json:"code"`
	Data struct {
		List []cloud139File `json:"list"`
		Page struct {
			NextPageCursor string `json:"nextPageCursor"`
			PageSize       int    `json:"pageSize"`
		} `json:"page"`
	} `json:"data"`
}

// cloud139DownloadResp 下载链接响应
type cloud139DownloadResp struct {
	Code string `json:"code"`
	Data struct {
		CdnURL string `json:"cdnUrl"`
		URL    string `json:"url"`
	} `json:"data"`
}

// Cloud139Provider 移动云盘提供者（personal_new 类型）
type Cloud139Provider struct {
	authorization  string // base64 编码的授权字符串
	account        string // 手机号
	token          string // 解析出的 token
	personalHost   string // 动态解析的个人云主机
	hostResolved   bool
	mu             sync.RWMutex
	client         *http.Client
}

// NewCloud139Provider 创建移动云盘提供者
// cookie 格式: base64 编码的 authorization 字符串
func NewCloud139Provider(cookie string) *Cloud139Provider {
	p := &Cloud139Provider{
		authorization: cookie,
		client:        internal.DefaultHTTPClient(),
	}
	p.parseAuthorization()
	return p
}

// parseAuthorization 解析 authorization 字符串
func (p *Cloud139Provider) parseAuthorization() {
	decoded, err := base64.StdEncoding.DecodeString(p.authorization)
	if err != nil {
		return
	}
	parts := strings.SplitN(string(decoded), ":", 3)
	if len(parts) >= 3 {
		p.account = parts[1]
		tokenParts := strings.Split(parts[2], "|")
		if len(tokenParts) > 0 {
			p.token = tokenParts[0]
		}
	}
}

func (p *Cloud139Provider) Name() string {
	return "139"
}

func (p *Cloud139Provider) Validate() (*cloudsync.CloudUserInfo, error) {
	if err := p.ensureReady(); err != nil {
		return nil, err
	}
	return &cloudsync.CloudUserInfo{
		Nickname: "移动云盘用户 (" + p.account + ")",
	}, nil
}

func (p *Cloud139Provider) ListFiles(pdirFid string, page, size int) ([]cloudsync.CloudFile, int, error) {
	if err := p.ensureReady(); err != nil {
		return nil, 0, err
	}

	if pdirFid == "" {
		pdirFid = "/"
	}

	// 139 云盘使用 cursor 分页
	var cursor string

	// 循环获取前面的页
	for i := 0; i < page; i++ {
		body := map[string]interface{}{
			"parentFileId": pdirFid,
			"pageInfo": map[string]interface{}{
				"pageSize":   size,
				"pageCursor": cursor,
			},
			"orderBy":        "updated_at",
			"orderDirection": "DESC",
		}

		var resp cloud139ListResp
		if err := p.post("/file/list", body, &resp); err != nil {
			return nil, 0, err
		}

		if resp.Code != "0" && resp.Code != "" {
			return nil, 0, fmt.Errorf("139云盘API错误: code=%s", resp.Code)
		}

		if i == page-1 {
			files := make([]cloudsync.CloudFile, 0, len(resp.Data.List))
			for _, f := range resp.Data.List {
				files = append(files, cloudsync.CloudFile{
					Fid:      f.FileID,
					FileName: f.Name,
					Size:     f.Size,
					IsDir:    f.Type == "folder",
				})
			}
			total := (page-1)*size + len(resp.Data.List)
			if resp.Data.Page.NextPageCursor != "" {
				total = page*size + 1
			}
			return files, total, nil
		}

		cursor = resp.Data.Page.NextPageCursor
		if cursor == "" {
			return nil, 0, nil
		}
	}

	return nil, 0, nil
}

func (p *Cloud139Provider) GetDownloadLink(fid string) (*cloudsync.CloudDownloadLink, error) {
	if err := p.ensureReady(); err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"fileId": fid,
	}

	var resp cloud139DownloadResp
	if err := p.post("/file/getDownloadUrl", body, &resp); err != nil {
		return nil, err
	}

	downloadURL := resp.Data.CdnURL
	if downloadURL == "" {
		downloadURL = resp.Data.URL
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("未获取到下载链接")
	}

	return &cloudsync.CloudDownloadLink{
		DownloadURL: downloadURL,
	}, nil
}

// ensureReady 确保 provider 已准备好（刷新 token + 解析主机）
func (p *Cloud139Provider) ensureReady() error {
	p.mu.RLock()
	if p.hostResolved && p.authorization != "" {
		p.mu.RUnlock()
		return nil
	}
	p.mu.RUnlock()

	// 尝试刷新 token
	if err := p.refreshToken(); err != nil {
		// 刷新失败不一定致命（可能 token 还有效）
	}

	// 解析个人云主机
	return p.resolvePersonalHost()
}

// refreshToken 刷新 authorization token
func (p *Cloud139Provider) refreshToken() error {
	decoded, err := base64.StdEncoding.DecodeString(p.authorization)
	if err != nil {
		return fmt.Errorf("authorization 解码失败: %w", err)
	}

	parts := strings.SplitN(string(decoded), ":", 3)
	if len(parts) < 3 {
		return fmt.Errorf("authorization 格式错误")
	}

	account := parts[1]
	tokenParts := strings.Split(parts[2], "|")
	if len(tokenParts) < 4 {
		return fmt.Errorf("token 格式错误，需要至少4个部分")
	}

	// 检查过期时间
	var expireMs int64
	fmt.Sscanf(tokenParts[3], "%d", &expireMs)
	if expireMs > 0 {
		remainDays := time.Until(time.UnixMilli(expireMs)).Hours() / 24
		if remainDays > 15 {
			// 还有超过15天，不需要刷新
			return nil
		}
	}

	// 发送 XML 刷新请求
	xmlBody := fmt.Sprintf(`<root><token>%s</token><account>%s</account><clienttype>656</clienttype></root>`,
		tokenParts[0], account)

	req, err := http.NewRequest("POST", cloud139TokenRefreshURL, strings.NewReader(xmlBody))
	if err != nil {
		return fmt.Errorf("创建刷新请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/xml")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("刷新token失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取刷新响应失败: %w", err)
	}

	var refreshResp cloud139TokenRefreshResp
	if err := xml.Unmarshal(respBytes, &refreshResp); err != nil {
		return fmt.Errorf("解析刷新响应失败: %w", err)
	}

	if refreshResp.Return != "0" || refreshResp.Token == "" {
		return fmt.Errorf("token刷新失败: %s", refreshResp.Desc)
	}

	// 重新编码 authorization
	newAuth := base64.StdEncoding.EncodeToString(
		[]byte(parts[0] + ":" + account + ":" + refreshResp.Token))

	p.mu.Lock()
	p.authorization = newAuth
	p.token = refreshResp.Token
	p.mu.Unlock()

	return nil
}

// resolvePersonalHost 解析个人云主机地址
func (p *Cloud139Provider) resolvePersonalHost() error {
	p.mu.RLock()
	if p.hostResolved {
		p.mu.RUnlock()
		return nil
	}
	p.mu.RUnlock()

	body := map[string]interface{}{
		"userInfo": map[string]interface{}{
			"userType":    1,
			"accountType": 1,
			"accountName": p.account,
		},
		"modAddrType": 1,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", cloud139RoutePolicyURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	p.setStandardHeaders(req, bodyBytes)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("获取路由策略失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取路由策略响应失败: %w", err)
	}

	var policyResp cloud139RoutePolicyResp
	if err := json.Unmarshal(respBytes, &policyResp); err != nil {
		return fmt.Errorf("解析路由策略失败: %w", err)
	}

	for _, policy := range policyResp.Data.RoutePolicyList {
		if policy.ModName == "personal" && policy.HttpsURL != "" {
			p.mu.Lock()
			p.personalHost = policy.HttpsURL
			p.hostResolved = true
			p.mu.Unlock()
			return nil
		}
	}

	return fmt.Errorf("未找到个人云主机地址")
}

// post 发送 POST 请求到个人云
func (p *Cloud139Provider) post(path string, body interface{}, result interface{}) error {
	p.mu.RLock()
	host := p.personalHost
	p.mu.RUnlock()

	if host == "" {
		return fmt.Errorf("个人云主机未解析")
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	req, err := http.NewRequest("POST", host+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	p.setStandardHeaders(req, bodyBytes)
	req.Header.Set("X-Yun-Api-Version", "v1")
	req.Header.Set("X-Yun-Module-Type", "100")
	req.Header.Set("X-Yun-Svc-Type", "1")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if result != nil {
		if err := json.Unmarshal(respBytes, result); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}
	return nil
}

// setStandardHeaders 设置标准请求头
func (p *Cloud139Provider) setStandardHeaders(req *http.Request, body []byte) {
	p.mu.RLock()
	auth := p.authorization
	p.mu.RUnlock()

	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("mcloud-channel", "1000101")
	req.Header.Set("mcloud-client", "10701")
	req.Header.Set("mcloud-version", "7.14.0")
	req.Header.Set("mcloud-sign", p.calSign(body))
	req.Header.Set("x-SvcType", "1")
	req.Header.Set("x-DeviceInfo", "||9|7.14.0|chrome|120.0.0.0|||windows 10||zh-CN|||")
	req.Header.Set("x-m4c-caller", "PC")
	req.Header.Set("x-m4c-src", "10002")
}

// calSign 计算请求签名
func (p *Cloud139Provider) calSign(body []byte) string {
	randStr := randomString(16)
	ts := time.Now().Format("2006-01-02 15:04:05")

	// URL 编码请求体
	encoded := url.QueryEscape(string(body))

	// 将编码后的字符串拆分为字符，排序，重新拼接
	chars := strings.Split(encoded, "")
	sort.Strings(chars)
	sorted := strings.Join(chars, "")

	// Base64 编码
	b64 := base64.StdEncoding.EncodeToString([]byte(sorted))

	// 签名
	part1 := md5Hex(b64)
	part2 := md5Hex(ts + ":" + randStr)
	sign := md5Hex(strings.ToUpper(part1 + ":" + part2))

	return ts + "," + randStr + "," + sign
}

// md5Hex 计算 MD5 十六进制字符串（小写）
func md5Hex(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

// randomString 生成随机字母数字字符串
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
