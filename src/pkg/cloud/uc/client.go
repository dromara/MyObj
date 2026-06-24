package uc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL = "https://pc-api.uc.cn"

	endpointShareToken  = "/1/clouddrive/share/sharepage/token"
	endpointShareDetail = "/1/clouddrive/share/sharepage/detail"
	endpointDownload    = "/1/clouddrive/file/download"
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

func (c *Client) GetShareToken(shareID, pwd string) (*ShareTokenResponse, error) {
	reqBody := ShareTokenRequest{
		ShareID: shareID,
		Pwd:     pwd,
	}

	var resp ShareTokenResponse
	if err := c.doRequest(http.MethodPost, endpointShareToken, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetShareDetail(shareID, shareToken, pwd string, page int) (*ShareDetailResponse, error) {
	reqBody := ShareDetailRequest{
		ShareID:    shareID,
		Pwd:        pwd,
		ShareToken: shareToken,
		Page:       page,
		Pr:         100,
		Scene:      "",
	}

	headers := map[string]string{
		"Cookie": "share_token=" + shareToken,
	}

	var resp ShareDetailResponse
	if err := c.doRequest(http.MethodPost, endpointShareDetail, reqBody, headers, &resp); err != nil {
		return nil, err
	}

	if resp.Status != 200 && resp.ErrNo != 0 {
		return nil, fmt.Errorf("API error: errno=%d, msg=%s", resp.ErrNo, resp.Msg)
	}

	return &resp, nil
}

func (c *Client) GetDownloadURL(shareID, shareToken string, fileIDs []string) (*DownloadURLResponse, error) {
	reqBody := DownloadURLRequest{
		ShareToken: shareToken,
		FileIDs:    fileIDs,
		ShareID:    shareID,
	}

	var resp DownloadURLResponse
	if err := c.doRequest(http.MethodPost, endpointDownload, reqBody, nil, &resp); err != nil {
		return nil, err
	}

	if resp.Status != 200 && resp.ErrNo != 0 {
		return nil, fmt.Errorf("API error: errno=%d, msg=%s", resp.ErrNo, resp.Msg)
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
		if err := json.Unmarshal(respBody, &apiResp); err == nil && apiResp.Msg != "" {
			return fmt.Errorf("API error: status=%d, errno=%d, msg=%s", apiResp.Status, apiResp.ErrNo, apiResp.Msg)
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
