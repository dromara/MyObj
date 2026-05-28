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
)

const (
	aliyunAPI = "https://openapi.alipan.com"
)

// 阿里云盘默认 OAuth 应用凭据（来自 alist）
const (
	aliyunDefaultCID   = ""
	aliyunDefaultCSEC  = ""
)

func init() {
	RegisterProvider("aliyun", func(cookie string) CloudProvider {
		return NewAliyunProvider(cookie)
	})
}

// aliyunErrResp 错误响应
type aliyunErrResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// aliyunFile 阿里云盘文件
type aliyunFile struct {
	DriveId      string `json:"drive_id"`
	FileId       string `json:"file_id"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	Type         string `json:"type"` // "file" or "folder"
	FileExtension string `json:"file_extension"`
}

// aliyunListResp 文件列表响应
type aliyunListResp struct {
	Items      []aliyunFile `json:"items"`
	NextMarker string       `json:"next_marker"`
}

// aliyunDownloadResp 下载链接响应
type aliyunDownloadResp struct {
	URL string `json:"url"`
}

// aliyunDriveInfo 云盘信息
type aliyunDriveInfo struct {
	TotalSize int64 `json:"total_size"`
	UsedSize  int64 `json:"used_size"`
}

// AliyunProvider 阿里云盘提供者（Open API）
type AliyunProvider struct {
	refreshToken string
	clientID     string
	clientSecret string
	accessToken  string
	driveID      string
	mu           sync.RWMutex
	client       *http.Client
}

// NewAliyunProvider 创建阿里云盘提供者
// cookie 格式: refresh_token 或 refresh_token|client_id|client_secret
func NewAliyunProvider(cookie string) *AliyunProvider {
	parts := strings.SplitN(cookie, "|", 3)
	refreshToken := parts[0]
	clientID := aliyunDefaultCID
	clientSecret := aliyunDefaultCSEC
	if len(parts) >= 3 && parts[1] != "" {
		clientID = parts[1]
		clientSecret = parts[2]
	}
	return &AliyunProvider{
		refreshToken: refreshToken,
		clientID:     clientID,
		clientSecret: clientSecret,
		client:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (a *AliyunProvider) Name() string {
	return "aliyun"
}

func (a *AliyunProvider) Validate() (*CloudUserInfo, error) {
	if err := a.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("刷新token失败: %w", err)
	}

	// 获取云盘信息
	var driveInfo aliyunDriveInfo
	if err := a.post("/adrive/v1.0/user/getDriveInfo", nil, &driveInfo); err != nil {
		return nil, err
	}

	// 获取用户信息
	var userInfo struct {
		NickName string `json:"nick_name"`
	}
	a.post("/adrive/v1.0/user/get", nil, &userInfo)

	nickname := userInfo.NickName
	if nickname == "" {
		nickname = "阿里云盘用户"
	}

	return &CloudUserInfo{
		Nickname:  nickname,
		TotalSize: driveInfo.TotalSize,
		UsedSize:  driveInfo.UsedSize,
	}, nil
}

func (a *AliyunProvider) ListFiles(pdirFid string, page, size int) ([]CloudFile, int, error) {
	if err := a.ensureAccessToken(); err != nil {
		return nil, 0, fmt.Errorf("刷新token失败: %w", err)
	}

	if pdirFid == "" {
		pdirFid = "root"
	}

	// 阿里云盘使用 marker 分页，需要跳过前面的页
	// 简单实现：每页请求 size 条，循环 page 次
	var marker string

	for i := 0; i < page; i++ {
		body := map[string]interface{}{
			"drive_id":        a.driveID,
			"parent_file_id":  pdirFid,
			"limit":           size,
			"fields":          "*",
			"order_by":        "updated_at",
			"order_direction": "DESC",
		}
		if marker != "" {
			body["marker"] = marker
		}

		var resp aliyunListResp
		if err := a.post("/adrive/v1.0/openFile/list", body, &resp); err != nil {
			return nil, 0, err
		}

		if i == page-1 {
			// 最后一页，返回结果
			files := make([]CloudFile, 0, len(resp.Items))
			for _, f := range resp.Items {
				files = append(files, CloudFile{
					Fid:      f.FileId,
					FileName: f.Name,
					Size:     f.Size,
					IsDir:    f.Type == "folder",
				})
			}
			total := (page-1)*size + len(resp.Items)
			if resp.NextMarker != "" {
				total = page*size + 1 // 可能还有更多
			}
			return files, total, nil
		}

		marker = resp.NextMarker
		if marker == "" {
			// 没有更多数据了
			return nil, 0, nil
		}
	}

	return nil, 0, nil
}

func (a *AliyunProvider) GetDownloadLink(fid string) (*CloudDownloadLink, error) {
	if err := a.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("刷新token失败: %w", err)
	}

	body := map[string]interface{}{
		"drive_id":  a.driveID,
		"file_id":   fid,
		"expire_sec": 14400,
	}

	var resp aliyunDownloadResp
	if err := a.post("/adrive/v1.0/openFile/getDownloadUrl", body, &resp); err != nil {
		return nil, err
	}

	if resp.URL == "" {
		return nil, fmt.Errorf("未获取到下载链接")
	}

	return &CloudDownloadLink{
		DownloadURL: resp.URL,
	}, nil
}

// ensureAccessToken 确保 access_token 有效
func (a *AliyunProvider) ensureAccessToken() error {
	a.mu.RLock()
	if a.accessToken != "" {
		a.mu.RUnlock()
		return nil
	}
	a.mu.RUnlock()
	return a.refreshAccessToken()
}

// refreshAccessToken 刷新 OAuth access_token
func (a *AliyunProvider) refreshAccessToken() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.accessToken != "" {
		return nil
	}

	if a.clientID == "" || a.clientSecret == "" {
		return fmt.Errorf("阿里云盘需要 client_id 和 client_secret，请在 cookie 字段提供: refresh_token|client_id|client_secret")
	}

	body := map[string]string{
		"refresh_token": a.refreshToken,
		"grant_type":    "refresh_token",
		"client_id":     a.clientID,
		"client_secret": a.clientSecret,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", aliyunAPI+"/oauth/access_token", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求token失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取token响应失败: %w", err)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(respBytes, &tokenResp); err != nil {
		return fmt.Errorf("解析token响应失败: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return fmt.Errorf("token刷新失败: access_token为空")
	}

	a.accessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		a.refreshToken = tokenResp.RefreshToken
	}

	// 获取 drive_id
	a.getDriveID()

	return nil
}

// getDriveID 获取默认云盘 ID
func (a *AliyunProvider) getDriveID() {
	var resp struct {
		DefaultDriveID string `json:"default_drive_id"`
	}
	if err := a.post("/adrive/v1.0/user/getDriveInfo", nil, &resp); err == nil {
		a.driveID = resp.DefaultDriveID
	}
}

// invalidateToken 使当前 access_token 失效
func (a *AliyunProvider) invalidateToken() {
	a.mu.Lock()
	a.accessToken = ""
	a.mu.Unlock()
}

// post 发送 POST 请求
func (a *AliyunProvider) post(path string, body interface{}, result interface{}) error {
	maxRetries := 2

	for attempt := 0; attempt < maxRetries; attempt++ {
		a.mu.RLock()
		token := a.accessToken
		a.mu.RUnlock()

		var reqBody io.Reader
		if body != nil {
			bodyBytes, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("序列化请求体失败: %w", err)
			}
			reqBody = bytes.NewReader(bodyBytes)
		} else {
			reqBody = bytes.NewReader([]byte("{}"))
		}

		req, err := http.NewRequest("POST", aliyunAPI+path, reqBody)
		if err != nil {
			return fmt.Errorf("创建请求失败: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := a.client.Do(req)
		if err != nil {
			return fmt.Errorf("请求失败: %w", err)
		}

		respBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("读取响应失败: %w", err)
		}

		// 检查错误
		var errResp aliyunErrResp
		if err := json.Unmarshal(respBytes, &errResp); err == nil && errResp.Code != "" {
			// Token 过期，刷新后重试
			if errResp.Code == "AccessTokenInvalid" || errResp.Code == "AccessTokenExpired" {
				a.invalidateToken()
				if err := a.refreshAccessToken(); err != nil {
					return fmt.Errorf("token刷新失败: %w", err)
				}
				continue
			}
			return fmt.Errorf("阿里云盘API错误: %s - %s", errResp.Code, errResp.Message)
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

