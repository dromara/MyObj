package caiyun

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL = "https://caiyun.139.com/portal/caiyun498"

	endpointGetShareInfo      = "/share/getShareInfo"
	endpointGetShareFileList  = "/share/getShareFileList"
	endpointGetShareDownloadURL = "/share/getShareDownloadUrl"
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

func (c *Client) GetShareInfo(shareID string) (*GetShareInfoResponse, error) {
	reqBody := GetShareInfoRequest{
		ShareID: shareID,
	}

	var resp GetShareInfoResponse
	if err := c.doRequest(http.MethodPost, endpointGetShareInfo, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	if resp.Code != "0" && resp.Code != "200" && resp.Code != "" {
		return nil, fmt.Errorf("API error: code=%s, message=%s", resp.Code, resp.Message)
	}

	return &resp, nil
}

func (c *Client) GetShareFileList(shareID, accessCode, catalogID string, pageNum, pageSize int) (*GetShareFileListResponse, error) {
	if pageSize <= 0 {
		pageSize = 200
	}
	if pageNum <= 0 {
		pageNum = 1
	}

	reqBody := GetShareFileListRequest{
		ShareID:    shareID,
		AccessCode: accessCode,
		CatalogID:  catalogID,
		PageNum:    pageNum,
		PageSize:   pageSize,
	}

	var resp GetShareFileListResponse
	if err := c.doRequest(http.MethodPost, endpointGetShareFileList, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	if resp.Code != "0" && resp.Code != "200" && resp.Code != "" {
		return nil, fmt.Errorf("API error: code=%s, message=%s", resp.Code, resp.Message)
	}

	return &resp, nil
}

func (c *Client) GetShareDownloadURL(shareID, accessCode, fileID string) (*GetShareDownloadURLResponse, error) {
	reqBody := GetShareDownloadURLRequest{
		ShareID:    shareID,
		AccessCode: accessCode,
		FileID:     fileID,
	}

	var resp GetShareDownloadURLResponse
	if err := c.doRequest(http.MethodPost, endpointGetShareDownloadURL, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	if resp.Code != "0" && resp.Code != "200" && resp.Code != "" {
		return nil, fmt.Errorf("API error: code=%s, message=%s", resp.Code, resp.Message)
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
	req.Header.Set("Referer", "https://caiyun.139.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

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
