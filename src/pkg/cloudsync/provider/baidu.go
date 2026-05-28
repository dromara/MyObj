package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/cloudsync/internal"
	"myobj/src/pkg/enum"
)

const (
	baiduPanAPI   = "https://pan.baidu.com/rest/2.0"
	baiduTokenAPI = "https://openapi.baidu.com/oauth/2.0/token"
	baiduCrackUA  = "netdisk"
)

func init() {
	cloudsync.Register(cloudsync.ProviderInfo{
		ID:          "baidu",
		Name:        "百度网盘",
		AuthType:    cloudsync.AuthRefreshToken,
		Description: "refresh_token 登录",
	}, enum.DownloadTaskTypeBaidu.Value(), func(cookie string) cloudsync.CloudProvider {
		return NewBaiduProvider(cookie)
	})
}

type baiduTokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

// baiduFile 百度网盘文件
type baiduFile struct {
	FsId           int64  `json:"fs_id"`
	Path           string `json:"path"`
	ServerFilename string `json:"server_filename"`
	Size           int64  `json:"size"`
	Isdir          int    `json:"isdir"`
	Md5            string `json:"md5"`
	Category       int    `json:"category"`
	ServerCtime    int64  `json:"server_ctime"`
	ServerMtime    int64  `json:"server_mtime"`
}

type baiduListResp struct {
	Errno int         `json:"errno"`
	List  []baiduFile `json:"list"`
}

type baiduDownloadResp struct {
	Errno int `json:"errno"`
	Info  []struct {
		Dlink string `json:"dlink"`
	} `json:"info"`
}

type baiduUserResp struct {
	Errno      int    `json:"errno"`
	BaiduName  string `json:"baidu_name"`
	VipType    int    `json:"vip_type"`
	TotalSpace int64  `json:"total_space"`
	UsedSpace  int64  `json:"used_space"`
}

// BaiduProvider 百度网盘提供者
type BaiduProvider struct {
	refreshToken string
	clientID     string
	clientSecret string
	accessToken  string
	mu           sync.RWMutex
	client       *http.Client
	credSync     credentialSyncHelper
}

// NewBaiduProvider 创建百度网盘提供者
func NewBaiduProvider(credential string) *BaiduProvider {
	cred := internal.ParseOAuthCredential(credential, cloudsync.BaiduClientID, cloudsync.BaiduClientSecret)
	return &BaiduProvider{
		refreshToken: cred.RefreshToken,
		clientID:     cred.ClientID,
		clientSecret: cred.ClientSecret,
		client:       internal.DefaultHTTPClient(),
	}
}

func (b *BaiduProvider) Name() string {
	return "baidu"
}

func (b *BaiduProvider) SetCredentialUpdateCallback(fn func(string)) {
	b.credSync.SetCredentialUpdateCallback(fn)
}

func (b *BaiduProvider) ExportCredential() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return internal.FormatOAuthCredential(b.refreshToken, b.clientID, b.clientSecret, cloudsync.BaiduClientID, cloudsync.BaiduClientSecret)
}

func (b *BaiduProvider) Validate() (*cloudsync.CloudUserInfo, error) {
	if err := b.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("刷新token失败: %w", err)
	}

	var resp baiduUserResp
	if err := b.get("/xpan/nas", map[string]string{"method": "uinfo"}, &resp); err != nil {
		return nil, err
	}

	if resp.Errno != 0 {
		return nil, fmt.Errorf("获取用户信息失败, errno: %d", resp.Errno)
	}

	return &cloudsync.CloudUserInfo{
		Nickname:  resp.BaiduName,
		TotalSize: resp.TotalSpace,
		UsedSize:  resp.UsedSpace,
	}, nil
}

func (b *BaiduProvider) ListFiles(pdirFid string, page, size int) ([]cloudsync.CloudFile, int, error) {
	if err := b.ensureAccessToken(); err != nil {
		return nil, 0, fmt.Errorf("刷新token失败: %w", err)
	}

	if pdirFid == "" {
		pdirFid = "/"
	}

	start := (page - 1) * size
	params := map[string]string{
		"method": "list",
		"dir":    pdirFid,
		"web":    "web",
		"start":  strconv.Itoa(start),
		"limit":  strconv.Itoa(size),
	}

	var resp baiduListResp
	if err := b.get("/xpan/file", params, &resp); err != nil {
		return nil, 0, err
	}

	if resp.Errno != 0 {
		return nil, 0, fmt.Errorf("获取文件列表失败, errno: %d", resp.Errno)
	}

	files := make([]cloudsync.CloudFile, 0, len(resp.List))
	for _, f := range resp.List {
		files = append(files, cloudsync.CloudFile{
			Fid:      f.Path,
			FileName: f.ServerFilename,
			Size:     f.Size,
			IsDir:    f.Isdir == 1,
		})
	}

	total := start + len(resp.List)
	if len(resp.List) == size {
		total = start + size + 1
	}

	return files, total, nil
}

func (b *BaiduProvider) GetDownloadLink(fid string) (*cloudsync.CloudDownloadLink, error) {
	if err := b.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("刷新token失败: %w", err)
	}

	params := map[string]string{
		"target": fmt.Sprintf("[\"%s\"]", fid),
		"dlink":  "1",
		"web":    "5",
		"origin": "dlna",
	}

	var resp baiduDownloadResp
	if err := b.request("https://pan.baidu.com/api/filemetas", "GET", params, nil, &resp); err != nil {
		return nil, err
	}

	if resp.Errno != 0 {
		return nil, fmt.Errorf("获取下载链接失败, errno: %d", resp.Errno)
	}

	if len(resp.Info) == 0 || resp.Info[0].Dlink == "" {
		return nil, fmt.Errorf("未获取到下载链接")
	}

	return &cloudsync.CloudDownloadLink{
		DownloadURL: resp.Info[0].Dlink,
		MustProxy:   true,
		Headers: map[string]string{
			"User-Agent": baiduCrackUA,
		},
	}, nil
}

func (b *BaiduProvider) ensureAccessToken() error {
	b.mu.RLock()
	if b.accessToken != "" {
		b.mu.RUnlock()
		return nil
	}
	b.mu.RUnlock()
	return b.refreshAccessToken()
}

func (b *BaiduProvider) refreshAccessToken() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.accessToken != "" {
		return nil
	}

	u := fmt.Sprintf("%s?grant_type=refresh_token&refresh_token=%s&client_id=%s&client_secret=%s",
		baiduTokenAPI, b.refreshToken, b.clientID, b.clientSecret)

	resp, err := b.client.Get(u)
	if err != nil {
		return fmt.Errorf("请求token失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取token响应失败: %w", err)
	}

	var tokenResp baiduTokenResp
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return fmt.Errorf("解析token响应失败: %w", err)
	}

	if tokenResp.Error != "" {
		return fmt.Errorf("token刷新失败: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	if tokenResp.RefreshToken == "" {
		return fmt.Errorf("token刷新失败: 返回的refresh_token为空")
	}

	b.accessToken = tokenResp.AccessToken
	b.refreshToken = tokenResp.RefreshToken
	b.credSync.notifyCredential(b.refreshToken, b.clientID, b.clientSecret, cloudsync.BaiduClientID, cloudsync.BaiduClientSecret)
	return nil
}

func (b *BaiduProvider) invalidateToken() {
	b.mu.Lock()
	b.accessToken = ""
	b.mu.Unlock()
}

func (b *BaiduProvider) get(path string, params map[string]string, result interface{}) error {
	return b.request(baiduPanAPI+path, "GET", params, nil, result)
}

func (b *BaiduProvider) request(url, method string, params map[string]string, body map[string]string, result interface{}) error {
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		b.mu.RLock()
		token := b.accessToken
		b.mu.RUnlock()

		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return fmt.Errorf("创建请求失败: %w", err)
		}

		q := req.URL.Query()
		q.Set("access_token", token)
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()

		if body != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}

		resp, err := b.client.Do(req)
		if err != nil {
			return fmt.Errorf("请求失败: %w", err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("读取响应失败: %w", err)
		}

		errno := extractErrno(bodyBytes)
		if errno == 0 {
			if result != nil {
				if err := json.Unmarshal(bodyBytes, result); err != nil {
					return fmt.Errorf("解析响应失败: %w", err)
				}
			}
			return nil
		}

		if errno == 111 || errno == -6 {
			b.invalidateToken()
			if err := b.refreshAccessToken(); err != nil {
				return fmt.Errorf("token刷新失败: %w", err)
			}
			continue
		}

		return fmt.Errorf("百度API错误, errno: %d, url: %s", errno, url)
	}

	return fmt.Errorf("请求重试%d次后仍失败", maxRetries)
}

func extractErrno(data []byte) int {
	var resp struct {
		Errno int `json:"errno"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return -1
	}
	return resp.Errno
}
