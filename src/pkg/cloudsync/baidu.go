package cloudsync

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	baiduPanAPI      = "https://pan.baidu.com/rest/2.0"
	baiduTokenAPI    = "https://openapi.baidu.com/oauth/2.0/token"
	baiduDefaultCID  = "hq9yQ9w9kR4YHj1kyYafLygVocobh7Sf"
	baiduDefaultCSEC = "YH2VpZcFJHYNnV6vLfHQXDBhcE7ZChyE"
	baiduCrackUA     = "netdisk"
)

func init() {
	RegisterProvider("baidu", func(cookie string) CloudProvider {
		return NewBaiduProvider(cookie)
	})
}

// baiduTokenResp OAuth token 刷新响应
type baiduTokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

// baiduFile 百度文件信息
type baiduFile struct {
	FsId           int64  `json:"fs_id"`
	Path           string `json:"path"`
	ServerFilename string `json:"server_filename"`
	Size           int64  `json:"size"`
	Isdir          int    `json:"isdir"` // 1=目录, 0=文件
	Md5            string `json:"md5"`
	Category       int    `json:"category"`
	ServerCtime    int64  `json:"server_ctime"`
	ServerMtime    int64  `json:"server_mtime"`
}

// baiduListResp 文件列表响应
type baiduListResp struct {
	Errno int         `json:"errno"`
	List  []baiduFile `json:"list"`
}

// baiduDownloadResp 下载链接响应（crack 方法）
type baiduDownloadResp struct {
	Errno int `json:"errno"`
	Info  []struct {
		Dlink string `json:"dlink"`
	} `json:"info"`
}

// baiduUserResp 用户信息响应
type baiduUserResp struct {
	Errno    int    `json:"errno"`
	BaiduName string `json:"baidu_name"`
	VipType  int    `json:"vip_type"` // 0=普通, 1=会员, 2=超级会员
	TotalSpace int64 `json:"total_space"`
	UsedSpace  int64 `json:"used_space"`
}

// BaiduProvider 百度网盘提供者
type BaiduProvider struct {
	refreshToken string
	clientID     string
	clientSecret string
	accessToken  string
	mu           sync.RWMutex
	client       *http.Client
}

// NewBaiduProvider 创建百度网盘提供者
// cookie 参数格式: refresh_token（直接传入百度 OAuth refresh_token）
func NewBaiduProvider(cookie string) *BaiduProvider {
	return &BaiduProvider{
		refreshToken: cookie,
		clientID:     baiduDefaultCID,
		clientSecret: baiduDefaultCSEC,
		client:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (b *BaiduProvider) Name() string {
		return "baidu"
}

func (b *BaiduProvider) Validate() (*CloudUserInfo, error) {
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

	return &CloudUserInfo{
		Nickname:  resp.BaiduName,
		TotalSize: resp.TotalSpace,
		UsedSize:  resp.UsedSpace,
	}, nil
}

func (b *BaiduProvider) ListFiles(pdirFid string, page, size int) ([]CloudFile, int, error) {
	if err := b.ensureAccessToken(); err != nil {
		return nil, 0, fmt.Errorf("刷新token失败: %w", err)
	}

	if pdirFid == "" {
		pdirFid = "/"
	}

	// 百度使用 start/limit 分页，需要从 page/size 转换
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

	files := make([]CloudFile, 0, len(resp.List))
	for _, f := range resp.List {
		files = append(files, CloudFile{
			Fid:      f.Path, // 百度使用路径作为文件标识
			FileName: f.ServerFilename,
			Size:     f.Size,
			IsDir:    f.Isdir == 1,
		})
	}

	// 百度 API 不直接返回总数，返回当前页数量作为近似值
	// 如果返回的数量等于请求的 size，可能还有更多
	total := start + len(resp.List)
	if len(resp.List) == size {
		total = start + size + 1 // 表示可能还有更多
	}

	return files, total, nil
}

func (b *BaiduProvider) GetDownloadLink(fid string) (*CloudDownloadLink, error) {
	if err := b.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("刷新token失败: %w", err)
	}

	// 使用 crack 方法获取下载链接（直接返回 dlink，无需跟随重定向）
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

	return &CloudDownloadLink{
		DownloadURL: resp.Info[0].Dlink,
		Headers: map[string]string{
			"User-Agent": baiduCrackUA,
		},
	}, nil
}

// ensureAccessToken 确保 access_token 有效，过期时自动刷新
func (b *BaiduProvider) ensureAccessToken() error {
	b.mu.RLock()
	if b.accessToken != "" {
		b.mu.RUnlock()
		return nil
	}
	b.mu.RUnlock()

	return b.refreshAccessToken()
}

// refreshAccessToken 刷新 OAuth access_token
func (b *BaiduProvider) refreshAccessToken() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 双重检查
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
	b.refreshToken = tokenResp.RefreshToken // 更新 refresh_token（百度会轮换）

	return nil
}

// invalidateToken 使当前 access_token 失效（用于重试）
func (b *BaiduProvider) invalidateToken() {
	b.mu.Lock()
	b.accessToken = ""
	b.mu.Unlock()
}

// get 发送 GET 请求
func (b *BaiduProvider) get(path string, params map[string]string, result interface{}) error {
	return b.request(baiduPanAPI+path, "GET", params, nil, result)
}

// request 发送请求并处理响应（带重试和 token 刷新）
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

		// 添加 access_token 作为查询参数
		q := req.URL.Query()
		q.Set("access_token", token)
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()

		if body != nil {
			// POST form 数据（当前未使用，预留）
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

		// 检查 errno
		errno := extractErrno(bodyBytes)
		if errno == 0 {
			// 成功
			if result != nil {
				if err := json.Unmarshal(bodyBytes, result); err != nil {
					return fmt.Errorf("解析响应失败: %w", err)
				}
			}
			return nil
		}

		// errno 111 或 -6 表示 token 过期，刷新后重试
		if errno == 111 || errno == -6 {
			b.invalidateToken()
			if err := b.refreshAccessToken(); err != nil {
				return fmt.Errorf("token刷新失败: %w", err)
			}
			continue
		}

		// 其他错误，不重试
		return fmt.Errorf("百度API错误, errno: %d, url: %s", errno, url)
	}

	return fmt.Errorf("请求重试%d次后仍失败", maxRetries)
}

// extractErrno 从 JSON 响应中提取 errno 字段
func extractErrno(data []byte) int {
	var resp struct {
		Errno int `json:"errno"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return -1 // 解析失败视为错误
	}
	return resp.Errno
}
