package sharelink

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	Register("baidu_share", parseBaiduShare)
}

var reBaiduSurl = regexp.MustCompile(`(?:pan\.baidu\.com/s/|yun\.baidu\.com/s/)([a-zA-Z0-9_-]+)`)

func parseBaiduShare(req ParseRequest) (*ParseResult, error) {
	cookie := strings.TrimSpace(req.Extra["cookie"])
	if cookie == "" {
		return nil, fmt.Errorf("百度分享需要 extra.cookie 中的 BDUSS")
	}
	surl := extractBaiduSurl(req.ShareURL)
	if surl == "" {
		return nil, fmt.Errorf("无效的百度分享链接")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	listURL := "https://pan.baidu.com/share/wxlist?channel=weixin&version=2.2.2&clienttype=25&web=1"
	form := url.Values{
		"shorturl": {surl},
		"pwd":      {req.Password},
		"dir":      {""},
		"root":     {"1"},
		"page":     {"1"},
		"num":      {"100"},
		"order":    {"time"},
	}
	httpReq, err := http.NewRequest(http.MethodPost, listURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("User-Agent", "netdisk")
	httpReq.Header.Set("Cookie", normalizeBDUSS(cookie))

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var parsed struct {
		Errno int `json:"errno"`
		Data  struct {
			List []struct {
				FsID           int64  `json:"fs_id"`
				ServerFilename string `json:"server_filename"`
				Size           int64  `json:"size"`
				IsDir          int    `json:"isdir"`
			} `json:"list"`
			ShareID int64  `json:"shareid"`
			UK      int64  `json:"uk"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.Errno != 0 || len(parsed.Data.List) == 0 {
		return nil, fmt.Errorf("百度分享列表失败 errno=%d", parsed.Errno)
	}
	file := parsed.Data.List[0]
	if file.IsDir == 1 {
		return nil, fmt.Errorf("分享根目录是文件夹，请使用文件分享链接")
	}

	dlink, fileName, size, err := baiduShareDlink(client, cookie, surl, file.FsID, parsed.Data.ShareID, parsed.Data.UK)
	if err != nil {
		return nil, err
	}
	if fileName == "" {
		fileName = file.ServerFilename
	}
	if size == 0 {
		size = file.Size
	}
	return &ParseResult{
		DownloadURL: dlink,
		FileName:    fileName,
		FileSize:    size,
		Headers: map[string]string{
			"User-Agent": "netdisk",
			"Cookie":     normalizeBDUSS(cookie),
		},
	}, nil
}

func baiduShareDlink(client *http.Client, cookie, surl string, fsID, shareID, uk int64) (string, string, int64, error) {
	cfgURL := "https://pan.baidu.com/share/tplconfig?fields=sign,timestamp&" + url.Values{"surl": {surl}}.Encode()
	cfgReq, _ := http.NewRequest(http.MethodGet, cfgURL, nil)
	cfgReq.Header.Set("User-Agent", "netdisk")
	cfgReq.Header.Set("Cookie", normalizeBDUSS(cookie))
	cfgResp, err := client.Do(cfgReq)
	if err != nil {
		return "", "", 0, err
	}
	cfgBody, _ := io.ReadAll(cfgResp.Body)
	cfgResp.Body.Close()
	var cfg struct {
		Errno int `json:"errno"`
		Data  struct {
			Sign      string `json:"sign"`
			Timestamp int64  `json:"timestamp"`
		} `json:"data"`
	}
	if err := json.Unmarshal(cfgBody, &cfg); err != nil {
		return "", "", 0, err
	}

	dlForm := url.Values{
		"fid_list":  {fmt.Sprintf("[%d]", fsID)},
		"primaryid": {strconv.FormatInt(shareID, 10)},
		"uk":        {strconv.FormatInt(uk, 10)},
		"type":      {"nolimit"},
		"product":   {"share"},
		"extra":     {`{"sekey":""}`},
	}
	dlURL := fmt.Sprintf("https://pan.baidu.com/api/sharedownload?app_id=250528&sign=%s&timestamp=%d", cfg.Data.Sign, cfg.Data.Timestamp)
	dlReq, _ := http.NewRequest(http.MethodPost, dlURL, strings.NewReader(dlForm.Encode()))
	dlReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	dlReq.Header.Set("User-Agent", "netdisk")
	dlReq.Header.Set("Cookie", normalizeBDUSS(cookie))
	dlResp, err := client.Do(dlReq)
	if err != nil {
		return "", "", 0, err
	}
	defer dlResp.Body.Close()
	dlBody, _ := io.ReadAll(dlResp.Body)
	var dl struct {
		Errno int `json:"errno"`
		List  []struct {
			Dlink          string `json:"dlink"`
			ServerFilename string `json:"server_filename"`
			Size           int64  `json:"size"`
		} `json:"list"`
	}
	if err := json.Unmarshal(dlBody, &dl); err != nil {
		return "", "", 0, err
	}
	if dl.Errno != 0 || len(dl.List) == 0 || dl.List[0].Dlink == "" {
		return "", "", 0, fmt.Errorf("百度分享下载失败 errno=%d", dl.Errno)
	}
	return dl.List[0].Dlink, dl.List[0].ServerFilename, dl.List[0].Size, nil
}

func extractBaiduSurl(raw string) string {
	m := reBaiduSurl.FindStringSubmatch(raw)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func normalizeBDUSS(cookie string) string {
	if strings.Contains(cookie, "BDUSS=") {
		return cookie
	}
	return "BDUSS=" + cookie
}
