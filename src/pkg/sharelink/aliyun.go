package sharelink

import (
	"bytes"
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
	Register("aliyun_share", parseAliyunShare)
}

var reAliShareID = regexp.MustCompile(`(?:/al/s/|/s/)([a-zA-Z0-9]+)`)

func parseAliyunShare(req ParseRequest) (*ParseResult, error) {
	refreshToken := strings.TrimSpace(req.Extra["refresh_token"])
	if refreshToken == "" {
		return nil, fmt.Errorf("阿里分享需要 extra.refresh_token 中的用户 refresh_token")
	}
	shareID := extractAliShareID(req.ShareURL)
	if shareID == "" {
		return nil, fmt.Errorf("无效的阿里云盘分享链接")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	accessToken, err := aliyunUserToken(client, refreshToken)
	if err != nil {
		return nil, err
	}
	shareToken, err := aliyunShareToken(client, accessToken, shareID, req.Password)
	if err != nil {
		return nil, err
	}

	listBody := map[string]interface{}{
		"share_id":        shareID,
		"parent_file_id":  "root",
		"limit":           200,
		"order_by":        "updated_at",
		"order_direction": "DESC",
	}
	listRaw, _ := json.Marshal(listBody)
	listReq, _ := http.NewRequest(http.MethodPost, "https://api.alipan.com/adrive/v3/file/list", bytes.NewReader(listRaw))
	setAliHeaders(listReq, accessToken, shareToken)
	listResp, err := client.Do(listReq)
	if err != nil {
		return nil, err
	}
	listBytes, _ := io.ReadAll(listResp.Body)
	listResp.Body.Close()

	var listParsed struct {
		Items []struct {
			FileID  string `json:"file_id"`
			Name    string `json:"name"`
			Type    string `json:"type"`
			Size    int64  `json:"size"`
			DriveID string `json:"drive_id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listBytes, &listParsed); err != nil {
		return nil, err
	}
	if len(listParsed.Items) == 0 {
		return nil, fmt.Errorf("阿里分享文件列表为空")
	}
	item := listParsed.Items[0]
	if item.Type == "folder" {
		return nil, fmt.Errorf("分享根目录是文件夹，请使用文件分享链接")
	}

	dlBody := map[string]interface{}{
		"drive_id":   item.DriveID,
		"file_id":    item.FileID,
		"share_id":   shareID,
		"expire_sec": 600,
	}
	dlRaw, _ := json.Marshal(dlBody)
	dlReq, _ := http.NewRequest(http.MethodPost, "https://api.alipan.com/v2/file/get_share_link_download_url", bytes.NewReader(dlRaw))
	setAliHeaders(dlReq, accessToken, shareToken)
	dlResp, err := client.Do(dlReq)
	if err != nil {
		return nil, err
	}
	dlBytes, _ := io.ReadAll(dlResp.Body)
	dlResp.Body.Close()
	var dlParsed struct {
		DownloadURL string `json:"download_url"`
	}
	if err := json.Unmarshal(dlBytes, &dlParsed); err != nil {
		return nil, err
	}
	if dlParsed.DownloadURL == "" {
		return nil, fmt.Errorf("未获取到阿里分享下载链接")
	}
	return &ParseResult{
		DownloadURL: dlParsed.DownloadURL,
		FileName:    item.Name,
		FileSize:    item.Size,
		Headers: map[string]string{
			"Referer": "https://www.alipan.com/",
		},
	}, nil
}

func aliyunUserToken(client *http.Client, refreshToken string) (string, error) {
	body := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}
	raw, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "https://auth.alipan.com/v2/account/token", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var parsed struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	if parsed.AccessToken == "" {
		return "", fmt.Errorf("阿里用户 token 获取失败")
	}
	return parsed.AccessToken, nil
}

func aliyunShareToken(client *http.Client, accessToken, shareID, pwd string) (string, error) {
	body := map[string]string{"share_id": shareID}
	if pwd != "" {
		body["share_pwd"] = pwd
	}
	raw, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "https://api.alipan.com/v2/share_link/get_share_token", bytes.NewReader(raw))
	setAliHeaders(req, accessToken, "")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var parsed struct {
		ShareToken string `json:"share_token"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	if parsed.ShareToken == "" {
		return "", fmt.Errorf("阿里分享 token 获取失败")
	}
	return parsed.ShareToken, nil
}

func setAliHeaders(req *http.Request, accessToken, shareToken string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if shareToken != "" {
		req.Header.Set("x-share-token", shareToken)
	}
	req.Header.Set("X-Canary", "client=web,app=share,version=v2.3.1")
}

func extractAliShareID(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	m := reAliShareID.FindStringSubmatch(u.Path)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}
