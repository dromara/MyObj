package xunlei

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL = "https://pan.xunlei.com"

	endpointShareLinkInfo = "/share/link/info"
	endpointShareFileList = "/share/link/list"
	endpointShareDownload = "/share/link/download"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetShareLinkInfo(shareID, pwd string) (*ShareLinkInfoResponse, error) {
	reqBody := ShareLinkInfoRequest{
		ShareID: shareID,
		Pwd:     pwd,
	}

	var resp ShareLinkInfoResponse
	if err := c.doRequest(http.MethodGet, endpointShareLinkInfo, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) ListShareFiles(shareID, parentFileID, pwd string, pageSize int) (*ShareFileListResponse, error) {
	if pageSize <= 0 {
		pageSize = 100
	}

	reqBody := ShareFileListRequest{
		ShareID:      shareID,
		ParentFileID: parentFileID,
		Pwd:          pwd,
		PageSize:     pageSize,
	}

	var resp ShareFileListResponse
	if err := c.doRequest(http.MethodGet, endpointShareFileList, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetShareDownload(shareID, fileID, pwd string) (*ShareDownloadResponse, error) {
	reqBody := ShareDownloadRequest{
		ShareID: shareID,
		FileID:  fileID,
		Pwd:     pwd,
	}

	var resp ShareDownloadResponse
	if err := c.doRequest(http.MethodGet, endpointShareDownload, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

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
