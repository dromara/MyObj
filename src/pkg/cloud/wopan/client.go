package wopan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL = "https://pan.wo.cn"

	endpointShareInfo   = "/api/share/getShareInfo"
	endpointVerifyPwd   = "/api/share/verifyPwd"
	endpointFileList    = "/api/share/getShareFileList"
	endpointDownloadURL = "/api/share/getDownloadUrl"
)

// Client 联通云盘API客户端
type Client struct {
	httpClient *http.Client
}

// NewClient 创建联通云盘客户端
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetShareInfo 获取分享信息
func (c *Client) GetShareInfo(shareID string) (*GetShareInfoResponse, error) {
	reqBody := GetShareInfoRequest{
		ShareID: shareID,
	}

	var resp GetShareInfoResponse
	if err := c.doRequest(http.MethodPost, endpointShareInfo, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: code=%d, message=%s", resp.Code, resp.Message)
	}

	return &resp, nil
}

// VerifySharePwd 验证分享密码
func (c *Client) VerifySharePwd(shareID, pwd string) (*VerifySharePwdResponse, error) {
	reqBody := VerifySharePwdRequest{
		ShareID: shareID,
		Pwd:     pwd,
	}

	var resp VerifySharePwdResponse
	if err := c.doRequest(http.MethodPost, endpointVerifyPwd, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: code=%d, message=%s", resp.Code, resp.Message)
	}

	return &resp, nil
}

// GetShareFileList 获取分享文件列表
func (c *Client) GetShareFileList(shareID, shareToken, parentFileID string, pageNum, pageSize int) (*GetShareFileListResponse, error) {
	if pageSize <= 0 {
		pageSize = 100
	}
	if pageNum <= 0 {
		pageNum = 1
	}

	reqBody := GetShareFileListRequest{
		ShareID:      shareID,
		ShareToken:   shareToken,
		ParentFileID: parentFileID,
		PageNum:      pageNum,
		PageSize:     pageSize,
	}

	var resp GetShareFileListResponse
	if err := c.doRequest(http.MethodPost, endpointFileList, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: code=%d, message=%s", resp.Code, resp.Message)
	}

	return &resp, nil
}

// GetDownloadURL 获取文件下载链接
func (c *Client) GetDownloadURL(shareID, shareToken, fileID string) (*GetDownloadURLResponse, error) {
	reqBody := GetDownloadURLRequest{
		ShareID:    shareID,
		ShareToken: shareToken,
		FileID:     fileID,
	}

	var resp GetDownloadURLResponse
	if err := c.doRequest(http.MethodPost, endpointDownloadURL, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: code=%d, message=%s", resp.Code, resp.Message)
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
		return fmt.Errorf("HTTP error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response failed: %w, body=%s", err, string(respBody))
		}
	}

	return nil
}
