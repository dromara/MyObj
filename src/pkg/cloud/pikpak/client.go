package pikpak

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL = "https://api-drive.mypikpak.com"

	endpointShareToken  = "/v1/sharing/invitation_code"
	endpointShareDetail = "/drive/v1/share/file/list"
	endpointDownloadURL = "/drive/v1/files"
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

func (c *Client) GetShareToken(shareID string) (*ShareTokenResponse, error) {
	reqBody := ShareTokenRequest{
		ShareID: shareID,
	}

	var resp ShareTokenResponse
	if err := c.doRequest(http.MethodPost, endpointShareToken, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetShareDetail(shareID, shareToken, parentID string, pageSize int) (*ShareDetailResponse, error) {
	if pageSize <= 0 {
		pageSize = 100
	}

	reqBody := ShareDetailRequest{
		ShareID:    shareID,
		ShareToken: shareToken,
		ParentID:   parentID,
		PageSize:   pageSize,
		ThumbnailSize: "SIZE_SMALL",
	}

	headers := map[string]string{
		"Authorization": "Bearer " + shareToken,
	}

	var resp ShareDetailResponse
	if err := c.doRequest(http.MethodPost, endpointShareDetail, reqBody, headers, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetDownloadURL(shareToken, fileID string) (*DownloadURLResponse, error) {
	endpoint := fmt.Sprintf("%s/%s?share_token=%s", endpointDownloadURL, fileID, shareToken)

	headers := map[string]string{
		"Authorization": "Bearer " + shareToken,
	}

	var resp DownloadURLResponse
	if err := c.doRequest(http.MethodGet, endpoint, nil, headers, &resp); err != nil {
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
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

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
