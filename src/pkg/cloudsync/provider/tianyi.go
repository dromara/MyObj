package provider

import (
	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/cloudsync/internal"
	
	"encoding/json"
	"fmt"
	"io"
	"myobj/src/pkg/enum"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	tianyiWebURL = "https://cloud.189.cn"
	tianyiUA     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	tianyiRootID = "-11"
)

func init() {
	cloudsync.Register(cloudsync.ProviderInfo{
		ID: "tianyi", Name: "天翼云盘", AuthType: cloudsync.AuthCookie,
		Description: "中国电信旗下，浏览器 Cookie 登录",
	}, enum.DownloadTaskTypeTianyi.Value(), func(cookie string) cloudsync.CloudProvider {
		return NewTianyiProvider(cookie)
	})
}

type tianyiFile struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type tianyiFolder struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type tianyiListResp struct {
	ResCode    int    `json:"res_code"`
	ResMessage string `json:"res_message"`
	FileListAO struct {
		Count      int            `json:"count"`
		FileList   []tianyiFile   `json:"fileList"`
		FolderList []tianyiFolder `json:"folderList"`
	} `json:"fileListAO"`
}

type tianyiUserResp struct {
	ResCode    int    `json:"res_code"`
	ResMessage string `json:"res_message"`
	Nickname   string `json:"nickname"`
	TotalSize  int64  `json:"totalSize"`
	UsedSize   int64  `json:"usedSize"`
}

type tianyiDownResp struct {
	ResCode     int    `json:"res_code"`
	ResMessage  string `json:"res_message"`
	DownloadURL string `json:"downloadUrl"`
}

type tianyiAPIError struct {
	Code    string `json:"errorCode"`
	Message string `json:"errorMsg"`
}

func (e *tianyiAPIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

// TianyiProvider 天翼云盘提供者（Web Cookie 认证）
type TianyiProvider struct {
	cookie string
	client *http.Client
}

func NewTianyiProvider(cookie string) *TianyiProvider {
	return &TianyiProvider{
		cookie: strings.TrimSpace(cookie),
		client: internal.DefaultHTTPClient(),
	}
}

func (p *TianyiProvider) Name() string { return "tianyi" }

func (p *TianyiProvider) Validate() (*cloudsync.CloudUserInfo, error) {
	var resp tianyiUserResp
	if err := p.getJSON(tianyiWebURL+"/v2/getUserBriefInfo.action", nil, &resp); err != nil {
		return nil, fmt.Errorf("Cookie无效或已过期: %w", err)
	}
	if resp.ResCode != 0 {
		return nil, fmt.Errorf("验证失败: %s", resp.ResMessage)
	}
	nickname := resp.Nickname
	if nickname == "" {
		nickname = "天翼云盘用户"
	}
	return &cloudsync.CloudUserInfo{
		Nickname:  nickname,
		TotalSize: resp.TotalSize,
		UsedSize:  resp.UsedSize,
	}, nil
}

func (p *TianyiProvider) ListFiles(pdirFid string, page, size int) ([]cloudsync.CloudFile, int, error) {
	if pdirFid == "" {
		pdirFid = tianyiRootID
	}
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 60
	}

	targetPage := page
	pageNum := 1
	var lastFiles []cloudsync.CloudFile

	for pageNum <= targetPage {
		q := url.Values{
			"folderId":   {pdirFid},
			"pageSize":   {strconv.Itoa(size)},
			"pageNum":    {strconv.Itoa(pageNum)},
			"mediaType":  {"0"},
			"iconOption": {"5"},
			"orderBy":    {"lastOpTime"},
			"descending": {"true"},
			"noCache":    {randomNoCache()},
		}

		var resp tianyiListResp
		if err := p.getJSON(tianyiWebURL+"/api/open/file/listFiles.action", q, &resp); err != nil {
			return nil, 0, err
		}
		if resp.ResCode != 0 {
			return nil, 0, fmt.Errorf("天翼云盘API错误: %s", resp.ResMessage)
		}

		files := make([]cloudsync.CloudFile, 0, len(resp.FileListAO.FolderList)+len(resp.FileListAO.FileList))
		for _, f := range resp.FileListAO.FolderList {
			files = append(files, cloudsync.CloudFile{
				Fid:      strconv.FormatInt(f.ID, 10),
				FileName: f.Name,
				IsDir:    true,
			})
		}
		for _, f := range resp.FileListAO.FileList {
			files = append(files, cloudsync.CloudFile{
				Fid:      strconv.FormatInt(f.ID, 10),
				FileName: f.Name,
				Size:     f.Size,
				IsDir:    false,
			})
		}

		if pageNum == targetPage {
			total := (page-1)*size + len(files)
			if resp.FileListAO.Count >= size {
				total = page*size + 1
			}
			return files, total, nil
		}

		lastFiles = files
		if resp.FileListAO.Count == 0 {
			break
		}
		pageNum++
	}

	return lastFiles, len(lastFiles), nil
}

func (p *TianyiProvider) GetDownloadLink(fid string) (*cloudsync.CloudDownloadLink, error) {
	q := url.Values{
		"fileId":  {fid},
		"noCache": {randomNoCache()},
	}
	var resp tianyiDownResp
	if err := p.getJSON(tianyiWebURL+"/api/portal/getFileInfo.action", q, &resp); err != nil {
		return nil, err
	}
	if resp.ResCode != 0 {
		return nil, fmt.Errorf("获取下载链接失败: %s", resp.ResMessage)
	}
	downloadURL := resp.DownloadURL
	if downloadURL == "" {
		return nil, fmt.Errorf("未获取到下载链接")
	}
	if strings.HasPrefix(downloadURL, "//") {
		downloadURL = "https:" + downloadURL
	}
	return &cloudsync.CloudDownloadLink{
		DownloadURL: downloadURL,
		Headers: map[string]string{
			"User-Agent": tianyiUA,
			"Referer":    tianyiWebURL + "/",
			"Cookie":     p.cookie,
		},
	}, nil
}

func (p *TianyiProvider) getJSON(apiURL string, query url.Values, dest interface{}) error {
	if query == nil {
		query = url.Values{}
	}
	if query.Get("noCache") == "" {
		query.Set("noCache", randomNoCache())
	}
	fullURL := apiURL
	if encoded := query.Encode(); encoded != "" {
		if strings.Contains(apiURL, "?") {
			fullURL += "&" + encoded
		} else {
			fullURL += "?" + encoded
		}
	}

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json;charset=UTF-8")
	req.Header.Set("User-Agent", tianyiUA)
	req.Header.Set("Referer", tianyiWebURL+"/")
	req.Header.Set("Cookie", p.cookie)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiErr tianyiAPIError
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Code != "" {
		if apiErr.Code == "InvalidSessionKey" {
			return fmt.Errorf("Cookie已过期，请重新获取")
		}
		return &apiErr
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	return nil
}

func randomNoCache() string {
	return strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
}
