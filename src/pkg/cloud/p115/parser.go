package p115

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"myobj/src/pkg/cloud"
)

var _ cloud.CloudProvider = (*P115Provider)(nil)

type P115Provider struct {
	client *Client
}

func NewP115Provider() *P115Provider {
	return &P115Provider{
		client: NewClient(),
	}
}

func (p *P115Provider) Name() string {
	return "115"
}

func (p *P115Provider) ParseShareLink(ctx context.Context, urlStr, pwd string) (*cloud.ShareInfo, error) {
	shareID, err := parse115ShareURL(urlStr)
	if err != nil {
		return nil, err
	}

	snapResp, err := p.client.GetShareSnap(shareID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "分享链接已失效") || strings.Contains(errMsg, "分享已过期") {
			return nil, cloud.ErrShareNotFound
		}
		if strings.Contains(errMsg, "提取码") || strings.Contains(errMsg, "密码") {
			return nil, cloud.ErrSharePasswordWrong
		}
		return nil, fmt.Errorf("get share snap failed: %w", err)
	}

	if snapResp.IsExpired {
		return nil, cloud.ErrShareNotFound
	}

	if snapResp.HasPassword && pwd == "" {
		return nil, cloud.ErrSharePasswordRequired
	}

	files := make([]cloud.ShareFile, 0, len(snapResp.Items))
	var totalSize int64
	for _, item := range snapResp.Items {
		files = append(files, convertSnapItem(item))
		totalSize += item.Size
	}

	var expiresAt *time.Time
	if snapResp.ExpiredTime > 0 {
		t := time.Unix(snapResp.ExpiredTime, 0)
		expiresAt = &t
	}

	return &cloud.ShareInfo{
		ShareID:   shareID,
		Title:     snapResp.ShareTitle,
		FileCount: snapResp.FileCount + snapResp.FolderCount,
		TotalSize: totalSize,
		Files:     files,
		ExpiresAt: expiresAt,
	}, nil
}

func (p *P115Provider) ListShareFiles(ctx context.Context, shareID, parentFileID string) ([]cloud.ShareFile, error) {
	if parentFileID == "" {
		snapResp, err := p.client.GetShareSnap(shareID)
		if err != nil {
			return nil, fmt.Errorf("get share snap failed: %w", err)
		}

		files := make([]cloud.ShareFile, 0, len(snapResp.Items))
		for _, item := range snapResp.Items {
			files = append(files, convertSnapItem(item))
		}
		return files, nil
	}

	dirResp, err := p.client.GetShareSnapDir(shareID, parentFileID)
	if err != nil {
		return nil, fmt.Errorf("get share snap dir failed: %w", err)
	}

	files := make([]cloud.ShareFile, 0, len(dirResp.Items))
	for _, item := range dirResp.Items {
		files = append(files, convertSnapItem(item))
	}
	return files, nil
}

func (p *P115Provider) GetDownloadLink(ctx context.Context, shareID, fileID string) (*cloud.DownloadInfo, error) {
	snapResp, err := p.client.GetShareSnap(shareID)
	if err != nil {
		return nil, fmt.Errorf("get share snap failed: %w", err)
	}

	var pickCode string
	for _, item := range snapResp.Items {
		if item.FileID == fileID {
			pickCode = item.PickCode
			break
		}
	}

	if pickCode == "" {
		dirResp, err := p.client.GetShareSnapDir(shareID, "0")
		if err == nil {
			for _, item := range dirResp.Items {
				if item.FileID == fileID {
					pickCode = item.PickCode
					break
				}
			}
		}
	}

	if pickCode == "" {
		return nil, cloud.ErrFileNotFound
	}

	downloadResp, err := p.client.GetShareDownloadURL(shareID, fileID, pickCode)
	if err != nil {
		return nil, fmt.Errorf("get share download url failed: %w", err)
	}

	if len(downloadResp.Urls) == 0 {
		return nil, fmt.Errorf("no download url available")
	}

	var downloadURL string
	for _, u := range downloadResp.Urls {
		if u.URL != "" {
			downloadURL = u.URL
			break
		}
	}

	if downloadURL == "" {
		return nil, fmt.Errorf("no valid download url found")
	}

	return &cloud.DownloadInfo{
		URL:        downloadURL,
		Size:       downloadResp.FileSize,
		Expiration: 15 * time.Minute,
		Headers:    map[string]string{},
	}, nil
}

func (p *P115Provider) DownloadFile(ctx context.Context, shareID, fileID string) (io.ReadCloser, error) {
	downloadInfo, err := p.GetDownloadLink(ctx, shareID, fileID)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadInfo.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create download request failed: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do download request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed: status=%d", resp.StatusCode)
	}

	return resp.Body, nil
}

func convertSnapItem(item ShareSnapItem) cloud.ShareFile {
	var updatedAt time.Time
	if item.UpdateTime > 0 {
		updatedAt = time.Unix(item.UpdateTime, 0)
	} else if item.CreateTime > 0 {
		updatedAt = time.Unix(item.CreateTime, 0)
	}

	return cloud.ShareFile{
		FileID:       item.FileID,
		Name:         item.Name,
		Size:         item.Size,
		IsDir:        item.IsDir,
		UpdatedAt:    updatedAt,
		Thumbnail:    item.ThumbURL,
		ParentFileID: item.ParentID,
	}
}

func parse115ShareURL(urlStr string) (string, error) {
	if !strings.Contains(urlStr, "://") {
		if len(urlStr) >= 5 {
			return urlStr, nil
		}
		return "", cloud.ErrInvalidShareURL
	}

	host := strings.ToLower(urlStr)
	if !strings.Contains(host, "115.com") && !strings.Contains(host, "115cdn.com") {
		return "", fmt.Errorf("%w: not a valid 115 URL", cloud.ErrInvalidShareURL)
	}

	idx := strings.Index(urlStr, "/s/")
	if idx < 0 {
		return "", cloud.ErrInvalidShareURL
	}

	shareID := urlStr[idx+3:]
	if qIdx := strings.Index(shareID, "?"); qIdx > 0 {
		shareID = shareID[:qIdx]
	}
	if qIdx := strings.Index(shareID, "#"); qIdx > 0 {
		shareID = shareID[:qIdx]
	}
	if shareID == "" {
		return "", cloud.ErrInvalidShareURL
	}

	return shareID, nil
}
