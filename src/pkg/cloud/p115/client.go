package p115

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL = "https://webapi.115.com"

	endpointShareSnap    = "/share/snap"
	endpointShareSnapDir = "/share/snap"
	endpointShareDownload = "/share/download"
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

func (c *Client) GetShareSnap(shareCode string) (*ShareSnapResponse, error) {
	reqURL := fmt.Sprintf("%s%s?share_code=%s", baseURL, endpointShareSnap, shareCode)

	var resp ShareSnapResponse
	if err := c.doRequest(http.MethodGet, reqURL, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetShareSnapDir(shareCode, dirID string) (*ShareSnapDirResponse, error) {
	reqURL := fmt.Sprintf("%s%s?share_code=%s&dir_id=%s", baseURL, endpointShareSnapDir, shareCode, dirID)

	var resp ShareSnapDirResponse
	if err := c.doRequest(http.MethodGet, reqURL, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetShareDownloadURL(shareCode, fileID, pickCode string) (*ShareDownloadResponse, error) {
	reqURL := fmt.Sprintf("%s%s?share_code=%s&file_id=%s&pick_code=%s",
		baseURL, endpointShareDownload, shareCode, fileID, pickCode)

	var resp ShareDownloadResponse
	if err := c.doRequest(http.MethodGet, reqURL, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) doRequest(method, reqURL string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

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
		if err := json.Unmarshal(respBody, &apiResp); err == nil && !apiResp.State {
			return fmt.Errorf("API error: errno=%d, error=%s", apiResp.ErrNo, apiResp.Error)
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
