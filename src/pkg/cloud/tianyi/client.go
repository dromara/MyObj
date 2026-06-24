package tianyi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL = "https://cloud.189.cn/api/portal"

	endpointGetShareInfo       = "/getShareInfoByCodeV2.action"
	endpointListShareDir       = "/listShareDirByShareId.action"
	endpointGetDownloadURL     = "/getShareDownloadUrl.action"
	endpointCheckAccessCode    = "/checkAccessCode.action"
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

func (c *Client) GetShareInfo(shareCode string) (*ShareInfoResponse, error) {
	params := map[string]string{
		"shareCode": shareCode,
	}
	var resp ShareInfoResponse
	if err := c.doRequest(http.MethodGet, endpointGetShareInfo, params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) CheckAccessCode(shareID, accessCode string) (*LoginShareResponse, error) {
	params := map[string]string{
		"shareId":    shareID,
		"accessCode": accessCode,
	}
	var resp LoginShareResponse
	if err := c.doRequest(http.MethodGet, endpointCheckAccessCode, params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ListShareDir(shareID string, parentFolderID string, pageNum, pageSize int) (*ShareFileListResponse, error) {
	if pageSize <= 0 {
		pageSize = 100
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	params := map[string]string{
		"shareId":   shareID,
		"pageNum":   fmt.Sprintf("%d", pageNum),
		"pageSize":  fmt.Sprintf("%d", pageSize),
	}
	if parentFolderID != "" {
		params["parentFolderId"] = parentFolderID
	}
	var resp ShareFileListResponse
	if err := c.doRequest(http.MethodGet, endpointListShareDir, params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetShareDownloadURL(shareID, fileID string) (*DownloadURLResponse, error) {
	params := map[string]string{
		"shareId": shareID,
		"fileId":  fileID,
	}
	var resp DownloadURLResponse
	if err := c.doRequest(http.MethodGet, endpointGetDownloadURL, params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) doRequest(method, endpoint string, params map[string]string, headers map[string]string, result interface{}) error {
	apiURL := baseURL + endpoint

	var reqBody io.Reader
	if method == http.MethodPost {
		formParts := make([]string, 0, len(params))
		for k, v := range params {
			formParts = append(formParts, k+"="+v)
		}
		reqBody = strings.NewReader(strings.Join(formParts, "&"))
	} else {
		if len(params) > 0 {
			queryParts := make([]string, 0, len(params))
			for k, v := range params {
				queryParts = append(queryParts, k+"="+v)
			}
			apiURL += "?" + strings.Join(queryParts, "&")
		}
	}

	req, err := http.NewRequest(method, apiURL, reqBody)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
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
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.ErrorCode != "" {
			return fmt.Errorf("%w: code=%s, msg=%s", ErrTianyiAPIError, apiErr.ErrorCode, apiErr.ErrorMsg)
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
