package cloudsync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// QuarkConfig 夸克/UC 网盘配置
type QuarkConfig struct {
	API     string // API 基地址
	Referer string // Referer 请求头
	UA      string // User-Agent 请求头
	PR      string // pr 查询参数
}

var quarkConf = QuarkConfig{
	API:     "https://drive.quark.cn/1/clouddrive",
	Referer: "https://pan.quark.cn",
	UA:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) quark-cloud-drive/2.5.20 Chrome/100.0.4896.160 Electron/18.3.5.4-b478491100 Safari/537.36 Channel/pckk_other_ch",
	PR:      "ucpro",
}

func init() {
	RegisterProvider("quark", func(cookie string) CloudProvider {
		return NewQuarkProviderWithConfig(cookie, quarkConf)
	})
}

// quarkResp 夸克 API 通用响应（仅用于错误检查）
type quarkResp struct {
	Status  int    `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// quarkFile 夸克文件信息
type quarkFile struct {
	Fid      string `json:"fid"`
	FileName string `json:"file_name"`
	Size     int64  `json:"size"`
	File     bool   `json:"file"` // true=文件, false=目录
}

// quarkFileListResp 文件列表完整响应
type quarkFileListResp struct {
	quarkResp
	Data struct {
		List []quarkFile `json:"list"`
	} `json:"data"`
	Metadata struct {
		Total int `json:"_total"`
	} `json:"metadata"`
}

// quarkDownResp 下载链接响应
type quarkDownResp struct {
	quarkResp
	Data []struct {
		DownloadURL string `json:"download_url"`
	} `json:"data"`
}

// QuarkProvider 夸克/UC 网盘提供者
type QuarkProvider struct {
	cookie string
	client *http.Client
	conf   QuarkConfig
}

// NewQuarkProviderWithConfig 创建夸克/UC 网盘提供者
func NewQuarkProviderWithConfig(cookie string, conf QuarkConfig) *QuarkProvider {
	return &QuarkProvider{
		cookie: cookie,
		client: &http.Client{Timeout: 30 * time.Second},
		conf:   conf,
	}
}

func (q *QuarkProvider) Name() string {
	return "quark"
}

func (q *QuarkProvider) ListFiles(pdirFid string, page, size int) ([]CloudFile, int, error) {
	if pdirFid == "" {
		pdirFid = "0"
	}

	params := url.Values{
		"pdir_fid":             {pdirFid},
		"_size":                {fmt.Sprintf("%d", size)},
		"_page":                {fmt.Sprintf("%d", page)},
		"_fetch_total":         {"1"},
		"fetch_all_file":       {"1"},
		"fetch_risk_file_name": {"1"},
		"_sort":                {"file_type:asc,updated_at:desc"},
	}

	var resp quarkFileListResp
	if err := q.get("/file/sort", params, &resp); err != nil {
		return nil, 0, err
	}

	files := make([]CloudFile, 0, len(resp.Data.List))
	for _, f := range resp.Data.List {
		files = append(files, CloudFile{
			Fid:      f.Fid,
			FileName: f.FileName,
			Size:     f.Size,
			IsDir:    !f.File,
		})
	}

	return files, resp.Metadata.Total, nil
}

func (q *QuarkProvider) GetDownloadLink(fid string) (*CloudDownloadLink, error) {
	body := map[string]interface{}{
		"fids": []string{fid},
	}

	var resp quarkDownResp
	if err := q.post("/file/download", body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("未获取到下载链接")
	}

	return &CloudDownloadLink{
		DownloadURL: resp.Data[0].DownloadURL,
		Headers: map[string]string{
			"Cookie":     q.cookie,
			"Referer":    q.conf.Referer,
			"User-Agent": q.conf.UA,
		},
	}, nil
}

func (q *QuarkProvider) Validate() (*CloudUserInfo, error) {
	var resp quarkResp
	if err := q.get("/config", nil, &resp); err != nil {
		return nil, fmt.Errorf("Cookie无效或已过期: %w", err)
	}

	return &CloudUserInfo{
		Nickname: "夸克用户",
	}, nil
}

// get 发送 GET 请求
func (q *QuarkProvider) get(path string, params url.Values, result interface{}) error {
	u := q.conf.API + path
	if params != nil {
		params.Set("pr", q.conf.PR)
		params.Set("fr", "pc")
		u += "?" + params.Encode()
	} else {
		u += "?" + "pr=" + q.conf.PR + "&fr=pc"
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	q.setHeaders(req)

	return q.doRequest(req, result)
}

// post 发送 POST 请求
func (q *QuarkProvider) post(path string, body interface{}, result interface{}) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	u := q.conf.API + path + "?pr=" + q.conf.PR + "&fr=pc"
	req, err := http.NewRequest("POST", u, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	q.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	return q.doRequest(req, result)
}

// setHeaders 设置通用请求头
func (q *QuarkProvider) setHeaders(req *http.Request) {
	req.Header.Set("Cookie", q.cookie)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", q.conf.Referer)
	req.Header.Set("User-Agent", q.conf.UA)
}

// doRequest 执行请求并处理响应
func (q *QuarkProvider) doRequest(req *http.Request, result interface{}) error {
	resp, err := q.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 捕获 __puus 轮换 token
	for _, c := range resp.Cookies() {
		if c.Name == "__puus" {
			q.cookie = setCookieValue(q.cookie, "__puus", c.Value)
			break
		}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	var apiResp quarkResp
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Status >= 400 || apiResp.Code != 0 {
		return fmt.Errorf("夸克API错误: %s", apiResp.Message)
	}

	if result != nil {
		if err := json.Unmarshal(bodyBytes, result); err != nil {
			return fmt.Errorf("解析响应数据失败: %w", err)
		}
	}

	return nil
}

// setCookieValue 更新 cookie 字符串中指定 key 的值
func setCookieValue(cookieStr, name, value string) string {
	cookies := parseCookies(cookieStr)
	found := false
	for i, c := range cookies {
		if c.Name == name {
			cookies[i].Value = value
			found = true
			break
		}
	}
	if !found {
		cookies = append(cookies, &http.Cookie{Name: name, Value: value})
	}
	return cookiesToString(cookies)
}

// parseCookies 解析 cookie 字符串
func parseCookies(s string) []*http.Cookie {
	header := http.Header{}
	header.Add("Cookie", s)
	req := &http.Request{Header: header}
	return req.Cookies()
}

// cookiesToString 序列化 cookies 为字符串
func cookiesToString(cookies []*http.Cookie) string {
	parts := make([]string, len(cookies))
	for i, c := range cookies {
		parts[i] = c.Name + "=" + c.Value
	}
	return strings.Join(parts, "; ")
}
