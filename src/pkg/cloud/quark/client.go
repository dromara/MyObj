package quark

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// 夸克网盘API基础URL
	baseURL = "https://drive-pc.quark.cn"

	// API端点
	endpointShareToken    = "/1/clouddrive/share/sharepage/token"
	endpointShareDetail   = "/1/clouddrive/share/sharepage/detail"
	endpointFileToken     = "/1/clouddrive/share/sharepage/download_token"
	endpointDownload      = "/1/clouddrive/file/download"
)

// Client 夸克网盘API客户端
type Client struct {
	httpClient *http.Client
}

// NewClient 创建夸克网盘客户端
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetShareToken 获取分享Token
func (c *Client) GetShareToken(shareID, pwd string) (*ShareTokenResponse, error) {
	reqBody := ShareTokenRequest{
		ShareID:  shareID,
		Passcode: pwd,
	}

	var resp ShareTokenResponse
	if err := c.doRequest(http.MethodPost, endpointShareToken, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetShareDetail 获取分享详情和文件列表
func (c *Client) GetShareDetail(shareID, shareToken, pwd string, page, size int) (*ShareDetailResponse, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 100
	}

	reqBody := ShareDetailRequest{
		ShareID:    shareID,
		ShareToken: shareToken,
		Password:   pwd,
		Page:       page,
		Size:       size,
	}

	headers := map[string]string{
		"cookie": "share_token=" + shareToken,
	}

	var resp ShareDetailResponse
	if err := c.doRequest(http.MethodPost, endpointShareDetail, reqBody, headers, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetDownloadToken 获取文件下载token
func (c *Client) GetDownloadToken(shareID, shareToken string, fileIDs []string) (*ShareFileTokenResponse, error) {
	reqBody := ShareFileTokenRequest{
		ShareID:    shareID,
		ShareToken: shareToken,
		FileIDs:    fileIDs,
	}

	headers := map[string]string{
		"cookie": "share_token=" + shareToken,
	}

	var resp ShareFileTokenResponse
	if err := c.doRequest(http.MethodPost, endpointFileToken, reqBody, headers, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetDownloadURL 获取文件下载链接
func (c *Client) GetDownloadURL(shareID, shareToken, downloadToken, fileID, formatType string) (*DownloadResponse, error) {
	reqBody := DownloadRequest{
		ShareID:       shareID,
		ShareToken:    shareToken,
		DownloadToken: downloadToken,
		FileID:        fileID,
		FormatType:    formatType,
	}

	headers := map[string]string{
		"cookie": "share_token=" + shareToken,
	}

	var resp DownloadResponse
	if err := c.doRequest(http.MethodPost, endpointDownload, reqBody, headers, &resp); err != nil {
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

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

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

	if resp.StatusCode != http.StatusOK {
		var apiResp APIResponse
		if err := json.Unmarshal(respBody, &apiResp); err == nil && apiResp.Code != 0 {
			return fmt.Errorf("API error: code=%d, message=%s", apiResp.Code, apiResp.Message)
		}
		return fmt.Errorf("HTTP error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response failed: %w, body=%s", err, string(respBody))
		}
	}

	return nil
}
