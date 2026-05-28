package sharelink

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func init() {
	Register("115_share", parse115Share)
}

var re115ShareCode = regexp.MustCompile(`115cdn\.com/s/([a-zA-Z0-9]+)`)

func parse115Share(req ParseRequest) (*ParseResult, error) {
	cookie := strings.TrimSpace(req.Extra["cookie"])
	if cookie == "" {
		return nil, fmt.Errorf("115分享需要 extra.cookie 中的115账号 Cookie")
	}
	shareCode, receiveCode := extract115ShareCodes(req.ShareURL, req.Password)
	if shareCode == "" {
		return nil, fmt.Errorf("无效的115分享链接")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	snapURL := "https://115cdn.com/webapi/share/snap?" + url.Values{
		"share_code":   {shareCode},
		"receive_code": {receiveCode},
		"cid":          {"0"},
		"limit":        {"32"},
		"offset":       {"0"},
		"format":       {"json"},
	}.Encode()
	snapReq, _ := http.NewRequest(http.MethodGet, snapURL, nil)
	snapReq.Header.Set("Cookie", cookie)
	snapReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	snapReq.Header.Set("Referer", fmt.Sprintf("https://115cdn.com/s/%s", shareCode))
	snapResp, err := client.Do(snapReq)
	if err != nil {
		return nil, err
	}
	snapBody, _ := io.ReadAll(snapResp.Body)
	snapResp.Body.Close()

	var snap struct {
		State bool `json:"state"`
		Data  struct {
			List []struct {
				Fid string `json:"fid"`
				N   string `json:"n"`
				S   int64  `json:"s"`
				Fc  string `json:"fc"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(snapBody, &snap); err != nil {
		return nil, err
	}
	if !snap.State || len(snap.Data.List) == 0 {
		return nil, fmt.Errorf("115分享快照获取失败")
	}
	file := snap.Data.List[0]
	if file.Fc == "0" {
		return nil, fmt.Errorf("分享根目录是文件夹")
	}

	downURL := "https://115cdn.com/webapi/share/downurl?" + url.Values{
		"share_code":   {shareCode},
		"receive_code": {receiveCode},
		"file_id":      {file.Fid},
		"dl":           {"1"},
	}.Encode()
	downReq, _ := http.NewRequest(http.MethodGet, downURL, nil)
	downReq.Header.Set("Cookie", cookie)
	downReq.Header.Set("User-Agent", snapReq.Header.Get("User-Agent"))
	downReq.Header.Set("Referer", snapReq.Header.Get("Referer"))
	downResp, err := client.Do(downReq)
	if err != nil {
		return nil, err
	}
	downBody, _ := io.ReadAll(downResp.Body)
	downResp.Body.Close()
	var down struct {
		State bool `json:"state"`
		Data  struct {
			URL struct {
				URL string `json:"url"`
			} `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(downBody, &down); err != nil {
		return nil, err
	}
	if !down.State || down.Data.URL.URL == "" {
		return nil, fmt.Errorf("115分享下载链接获取失败")
	}
	return &ParseResult{
		DownloadURL: down.Data.URL.URL,
		FileName:    file.N,
		FileSize:    file.S,
		Headers: map[string]string{
			"User-Agent": snapReq.Header.Get("User-Agent"),
			"Referer":    snapReq.Header.Get("Referer"),
			"Cookie":     cookie,
		},
	}, nil
}

func extract115ShareCodes(raw, password string) (shareCode, receiveCode string) {
	m := re115ShareCode.FindStringSubmatch(raw)
	if len(m) >= 2 {
		shareCode = m[1]
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err == nil {
		if rc := u.Query().Get("password"); rc != "" {
			receiveCode = rc
		}
	}
	if receiveCode == "" {
		receiveCode = password
	}
	return shareCode, receiveCode
}
