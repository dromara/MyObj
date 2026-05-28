package cloudsync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	xunleiAPI     = "https://x-api-pan.xunlei.com/drive/v1"
	xunleiUserAPI = "https://xluser-ssl.xunlei.com/v1"
)

// 迅雷浏览器版默认凭据
const (
	xunleiDefaultCID  = "ZUBzD9J_XPXfn7f7"
	xunleiDefaultCSEC = "yESVmHecEe6F0aou69vl-g"
	xunleiDefaultUA   = "AndroidDownloadManager/13 (Linux; U; Android 13; M2004J7AC Build/SP1A.210812.016)"
)

func init() {
	RegisterProvider("xunlei", func(cookie string) CloudProvider {
		return NewXunleiProvider(cookie)
	})
}

// xunleiFile 迅雷文件
type xunleiFile struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Size           string `json:"size"` // 字符串类型
	WebContentLink string `json:"web_content_link"`
	Kind           string `json:"kind"` // "drive#file" 表示文件
	Trashed        bool   `json:"trashed"`
	ParentID       string `json:"parent_id"`
}

// xunleiListResp 文件列表响应
type xunleiListResp struct {
	Files      []xunleiFile `json:"files"`
	NextPageToken string    `json:"next_page_token"`
}

// xunleiTokenResp token 响应
type xunleiTokenResp struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	UserID       string `json:"user_id"`
}

// xunleiUserInfo 用户信息
type xunleiUserInfo struct {
	Name   string `json:"name"`
	Sub    string `json:"sub"`
	Space  string `json:"space"` // 空间大小（可能不存在）
}

// XunleiProvider 迅雷网盘提供者
type XunleiProvider struct {
	refreshToken string
	clientID     string
	clientSecret string
	accessToken  string
	tokenType    string
	deviceID     string
	mu           sync.RWMutex
	client       *http.Client
}

// NewXunleiProvider 创建迅雷网盘提供者
// cookie 格式: refresh_token 或 refresh_token|client_id|client_secret
func NewXunleiProvider(cookie string) *XunleiProvider {
	parts := strings.SplitN(cookie, "|", 3)
	refreshToken := parts[0]
	clientID := xunleiDefaultCID
	clientSecret := xunleiDefaultCSEC
	if len(parts) >= 3 && parts[1] != "" {
		clientID = parts[1]
		clientSecret = parts[2]
	}
	return &XunleiProvider{
		refreshToken: refreshToken,
		clientID:     clientID,
		clientSecret: clientSecret,
		deviceID:     uuid.New().String(),
		client:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (x *XunleiProvider) Name() string {
	return "xunlei"
}

func (x *XunleiProvider) Validate() (*CloudUserInfo, error) {
	if err := x.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("刷新token失败: %w", err)
	}

	var userInfo xunleiUserInfo
	if err := x.get(xunleiUserAPI+"/user/me", nil, &userInfo); err != nil {
		return nil, err
	}

	nickname := userInfo.Name
	if nickname == "" {
		nickname = "迅雷用户"
	}

	return &CloudUserInfo{
		Nickname: nickname,
	}, nil
}

func (x *XunleiProvider) ListFiles(pdirFid string, page, size int) ([]CloudFile, int, error) {
	if err := x.ensureAccessToken(); err != nil {
		return nil, 0, fmt.Errorf("刷新token失败: %w", err)
	}

	if pdirFid == "" {
		pdirFid = "" // 根目录
	}

	// 迅雷使用 page_token 分页
	// 简单实现：循环获取直到目标页
	var pageToken string

	for i := 0; i < page; i++ {
		params := map[string]string{
			"parent_id": pdirFid,
			"limit":     fmt.Sprintf("%d", size),
			"filters":   `{"trashed":{"eq":false}}`,
		}
		if pageToken != "" {
			params["page_token"] = pageToken
		}

		var resp xunleiListResp
		if err := x.get(xunleiAPI+"/files", params, &resp); err != nil {
			return nil, 0, err
		}

		if i == page-1 {
			files := make([]CloudFile, 0, len(resp.Files))
			for _, f := range resp.Files {
				if f.Trashed {
					continue
				}
				files = append(files, CloudFile{
					Fid:      f.ID,
					FileName: f.Name,
					Size:     parseSize(f.Size),
					IsDir:    f.Kind == "drive#folder",
				})
			}
			total := (page-1)*size + len(resp.Files)
			if resp.NextPageToken != "" {
				total = page*size + 1
			}
			return files, total, nil
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			return nil, 0, nil
		}
	}

	return nil, 0, nil
}

func (x *XunleiProvider) GetDownloadLink(fid string) (*CloudDownloadLink, error) {
	if err := x.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("刷新token失败: %w", err)
	}

	// 获取文件详情（包含下载链接）
	var file xunleiFile
	if err := x.get(xunleiAPI+"/files/"+fid, nil, &file); err != nil {
		return nil, err
	}

	if file.WebContentLink == "" {
		return nil, fmt.Errorf("未获取到下载链接")
	}

	return &CloudDownloadLink{
		DownloadURL: file.WebContentLink,
		Headers: map[string]string{
			"User-Agent": xunleiDefaultUA,
		},
	}, nil
}

// ensureAccessToken 确保 access_token 有效
func (x *XunleiProvider) ensureAccessToken() error {
	x.mu.RLock()
	if x.accessToken != "" {
		x.mu.RUnlock()
		return nil
	}
	x.mu.RUnlock()
	return x.refreshAccessToken()
}

// refreshAccessToken 刷新 access_token
func (x *XunleiProvider) refreshAccessToken() error {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.accessToken != "" {
		return nil
	}

	body := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": x.refreshToken,
		"client_id":     x.clientID,
		"client_secret": x.clientSecret,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", xunleiUserAPI+"/auth/token", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := x.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求token失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取token响应失败: %w", err)
	}

	var tokenResp xunleiTokenResp
	if err := json.Unmarshal(respBytes, &tokenResp); err != nil {
		return fmt.Errorf("解析token响应失败: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return fmt.Errorf("token刷新失败: access_token为空")
	}

	x.accessToken = tokenResp.AccessToken
	x.tokenType = tokenResp.TokenType
	if tokenResp.RefreshToken != "" {
		x.refreshToken = tokenResp.RefreshToken
	}

	return nil
}

// invalidateToken 使当前 access_token 失效
func (x *XunleiProvider) invalidateToken() {
	x.mu.Lock()
	x.accessToken = ""
	x.mu.Unlock()
}

// get 发送 GET 请求
func (x *XunleiProvider) get(url string, params map[string]string, result interface{}) error {
	maxRetries := 2

	for attempt := 0; attempt < maxRetries; attempt++ {
		x.mu.RLock()
		token := x.accessToken
		tokenType := x.tokenType
		x.mu.RUnlock()

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("创建请求失败: %w", err)
		}

		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()

		// 设置认证和设备头
		req.Header.Set("Authorization", tokenType+" "+token)
		req.Header.Set("User-Agent", xunleiDefaultUA)
		req.Header.Set("Accept", "application/json;charset=UTF-8")
		req.Header.Set("x-device-id", x.deviceID)
		req.Header.Set("x-client-id", x.clientID)
		req.Header.Set("x-client-version", "1.10.0.2633")

		resp, err := x.client.Do(req)
		if err != nil {
			return fmt.Errorf("请求失败: %w", err)
		}

		respBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("读取响应失败: %w", err)
		}

		// 检查错误码
		errCode := extractJSONInt(respBytes, "error_code")
		if errCode != 0 {
			// Token 相关错误，刷新后重试
			if errCode == 4122 || errCode == 4121 || errCode == 10 || errCode == 16 {
				x.invalidateToken()
				if err := x.refreshAccessToken(); err != nil {
					return fmt.Errorf("token刷新失败: %w", err)
				}
				continue
			}
			errMsg := extractJSONString(respBytes, "error_description")
			if errMsg == "" {
				errMsg = extractJSONString(respBytes, "message")
			}
			return fmt.Errorf("迅雷API错误: error_code=%d, %s", errCode, errMsg)
		}

		if result != nil {
			if err := json.Unmarshal(respBytes, result); err != nil {
				return fmt.Errorf("解析响应失败: %w", err)
			}
		}
		return nil
	}

	return fmt.Errorf("请求重试%d次后仍失败", maxRetries)
}

// extractJSONInt 从 JSON 中提取整数字段
func extractJSONInt(data []byte, key string) int {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return 0
	}
	if v, ok := m[key]; ok {
		var n int
		json.Unmarshal(v, &n)
		return n
	}
	return 0
}

// extractJSONString 从 JSON 中提取字符串字段
func extractJSONString(data []byte, key string) string {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	if v, ok := m[key]; ok {
		var s string
		json.Unmarshal(v, &s)
		return s
	}
	return ""
}

// parseSize 解析字符串类型的文件大小
func parseSize(s string) int64 {
	var size int64
	fmt.Sscanf(s, "%d", &size)
	return size
}
