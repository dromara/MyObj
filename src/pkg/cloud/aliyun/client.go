package aliyun

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// 阿里云盘API基础URL
	baseURL = "https://api.alipan.com"
	
	// API端点
	endpointShareToken         = "/v2/share_link/get_share_token"
	endpointShareLinkInfo      = "/v2/share_link/get_share_link_info"
	endpointShareFileList      = "/adrive/v3/file/list"  // 使用正确的文件列表端点
	endpointShareDownloadURL   = "/v2/file/get_share_link_download_url"
	
	// Canary Header 用于解除速率限制
	canaryHeaderKey   = "X-Canary"
	canaryHeaderValue = "client=web,app=share,version=v2.3.1"
)

// Client 阿里云盘API客户端
type Client struct {
	httpClient *http.Client
	appID      string
}

// NewClient 创建阿里云盘客户端
// appID: 应用ID（可选，留空使用默认值）
func NewClient(appID string) *Client {
	if appID == "" {
		appID = "25d4b43c40bc4a2b93a5b9d7d5d3c3f7" // 默认AppID
	}
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		appID: appID,
	}
}

// GetShareToken 获取分享Token
func (c *Client) GetShareToken(shareID, sharePwd string) (*ShareTokenResponse, error) {
	reqBody := ShareTokenRequest{
		ShareID: shareID,
		SharePwd: sharePwd,
	}
	
	var resp ShareTokenResponse
	if err := c.doRequest(http.MethodPost, endpointShareToken, reqBody, nil, &resp); err != nil {
		return nil, err
	}
	
	return &resp, nil
}

// GetShareLinkInfo 获取分享链接信息
func (c *Client) GetShareLinkInfo(shareID, shareToken string) (*ShareLinkInfoResponse, error) {
	reqBody := ShareLinkInfoRequest{
		ShareID: shareID,
	}
	
	headers := map[string]string{
		"x-share-token": shareToken,
	}
	
	var resp ShareLinkInfoResponse
	if err := c.doRequest(http.MethodPost, endpointShareLinkInfo, reqBody, headers, &resp); err != nil {
		return nil, err
	}
	
	return &resp, nil
}

// ListShareFiles 列出分享文件
func (c *Client) ListShareFiles(shareID, shareToken, parentFileID string, limit int) (*ShareFileListResponse, error) {
	if limit <= 0 {
		limit = 200
	}
	
	if parentFileID == "" {
		parentFileID = "root"
	}
	
	reqBody := map[string]interface{}{
		"share_id":                 shareID,
		"parent_file_id":          parentFileID,
		"limit":                   limit,
		"order_by":                "name",
		"order_direction":         "ASC",
		"image_thumbnail_process": "image/resize,w_160/format,jpeg",
		"image_url_process":       "image/resize,w_1920/format,jpeg",
		"video_thumbnail_process": "video/snapshot,t_1000,f_jpg,ar_auto,w_300",
	}
	
	headers := map[string]string{
		"x-share-token": shareToken,
		canaryHeaderKey: canaryHeaderValue,
	}
	
	var resp ShareFileListResponse
	if err := c.doRequest(http.MethodPost, endpointShareFileList, reqBody, headers, &resp); err != nil {
		return nil, err
	}
	
	return &resp, nil
}

// GetShareLinkDownloadURL 获取文件下载链接
func (c *Client) GetShareLinkDownloadURL(shareID, shareToken, fileID, driveID string) (*GetShareLinkDownloadURLResponse, error) {
	reqBody := map[string]interface{}{
		"share_id":    shareID,
		"file_id":     fileID,
		"expire_sec":  600,
		"share_token": shareToken,
	}
	if driveID != "" {
		reqBody["drive_id"] = driveID
	}

	headers := map[string]string{
		"x-share-token": shareToken,
		canaryHeaderKey: canaryHeaderValue,
	}

	var resp GetShareLinkDownloadURLResponse
	if err := c.doRequest(http.MethodPost, endpointShareDownloadURL, reqBody, headers, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetDownloadURLWithToken 使用access_token获取下载链接
func (c *Client) GetDownloadURLWithToken(accessToken, driveID, fileID string) (*GetShareLinkDownloadURLResponse, error) {
	reqBody := map[string]interface{}{
		"drive_id":   driveID,
		"file_id":    fileID,
		"expire_sec": 14400,
	}

	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
	}

	var resp GetShareLinkDownloadURLResponse
	if err := c.doRequest(http.MethodPost, "/adrive/v1.0/openFile/getDownloadUrl", reqBody, headers, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(method, endpoint string, body interface{}, headers map[string]string, result interface{}) error {
	url := baseURL + endpoint
	
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
		reqBody = strings.NewReader(string(jsonData))
	}
	
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}
	
	// 设置默认请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	// 设置自定义请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}
	
	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		var apiResp APIResponse
		if err := json.Unmarshal(respBody, &apiResp); err == nil && apiResp.Code != "" {
			return fmt.Errorf("API error: code=%s, message=%s", apiResp.Code, apiResp.Message)
		}
		return fmt.Errorf("HTTP error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}
	
	// 解析响应
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response failed: %w, body=%s", err, string(respBody))
		}
	}
	
	return nil
}
