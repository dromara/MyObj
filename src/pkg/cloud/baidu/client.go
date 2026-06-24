package baidu

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// 百度网盘API基础URL
	baseURL = "https://pan.baidu.com"

	// API端点
	endpointShareVerify   = "/share/verify"
	endpointShareList     = "/share/list"
	endpointShareDownload = "/share/download"
)

// Client 百度网盘API客户端
type Client struct {
	httpClient *http.Client
}

// NewClient 创建百度网盘客户端
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ShareVerify 验证分享链接（获取分享ID和UK）
func (c *Client) ShareVerify(surl, pwd string) (*ShareVerifyResponse, error) {
	reqBody := ShareVerifyRequest{
		Surl: surl,
		Pwd:  pwd,
	}

	var resp ShareVerifyResponse
	if err := c.doRequest(http.MethodPost, endpointShareVerify, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	if resp.Errno != 0 {
		return nil, fmt.Errorf("share verify failed: errno=%d, errmsg=%s", resp.Errno, resp.ErrMsg)
	}

	return &resp, nil
}

// ListShareFiles 获取分享文件列表
func (c *Client) ListShareFiles(shareID, uk string, page, num int) (*ShareFileListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if num <= 0 {
		num = 100
	}

	reqBody := ShareFileListRequest{
		ShareID: shareID,
		Uk:      uk,
		Page:    page,
		Num:     num,
		Order:   "time",
	}

	var resp ShareFileListResponse
	if err := c.doRequest(http.MethodPost, endpointShareList, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	if resp.Errno != 0 {
		return nil, fmt.Errorf("list share files failed: errno=%d, errmsg=%s", resp.Errno, resp.ErrMsg)
	}

	return &resp, nil
}

// GetShareDownloadURL 获取文件下载链接
func (c *Client) GetShareDownloadURL(shareID, uk string, fsID int64) (*ShareDownloadResponse, error) {
	reqBody := ShareDownloadRequest{
		ShareID: shareID,
		Uk:      uk,
		FsID:    fsID,
	}

	var resp ShareDownloadResponse
	if err := c.doRequest(http.MethodPost, endpointShareDownload, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	if resp.Errno != 0 {
		return nil, fmt.Errorf("get download url failed: errno=%d, errmsg=%s", resp.Errno, resp.ErrMsg)
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
		if err := json.Unmarshal(respBody, &apiResp); err == nil && apiResp.Errno != 0 {
			return fmt.Errorf("API error: errno=%d, errmsg=%s", apiResp.Errno, apiResp.ErrMsg)
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