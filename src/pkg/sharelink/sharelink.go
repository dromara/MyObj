package sharelink

import "fmt"

// ParseRequest share link parse input.
type ParseRequest struct {
	ShareURL string
	Password string
	// Extra optional credentials: cookie, refresh_token, access_token, etc.
	Extra map[string]string
}

// ParseResult share link parse output.
type ParseResult struct {
	DownloadURL  string            `json:"download_url"`
	FileName     string            `json:"file_name"`
	FileSize     int64             `json:"file_size"`
	FileSizeText string            `json:"file_size_text,omitempty"`
	Headers      map[string]string `json:"-"`
}

type parserFunc func(req ParseRequest) (*ParseResult, error)

var registry = map[string]parserFunc{}

// Register registers a share-link parser by provider id.
func Register(providerID string, fn parserFunc) {
	registry[providerID] = fn
}

// Parse parses a share link for the given provider.
func Parse(providerID string, req ParseRequest) (*ParseResult, error) {
	fn, ok := registry[providerID]
	if !ok {
		return nil, fmt.Errorf("不支持的分享类型: %s", providerID)
	}
	return fn(req)
}

// Supported returns registered share provider ids.
func Supported() []string {
	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	return ids
}
