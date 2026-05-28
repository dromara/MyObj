package lanzou

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var lanzouDomains = []string{
	"lanzoui.com", "lanzoux.com", "lanzouh.com", "lanzoub.com",
	"lanzouc.com", "lanzoup.com", "lanzouo.com", "lanzoum.com",
	"lanzouv.com", "lanzouy.com", "lanzouq.com", "lanzous.com",
	"lanzout.com", "lanzouu.com", "lanzouw.com", "lanzoue.com",
	"lanzouf.com", "lanzoug.com", "lanzouj.com", "lanzouk.com",
	"lanzoul.com", "lanzoun.com", "lanzour.com", "ilanzou.com",
	"lanzou.com", "lanzou.net",
}

// ParseResult 蓝奏云解析结果
type ParseResult struct {
	DownloadURL  string `json:"download_url"`
	FileName     string `json:"file_name"`
	FileSize     int64  `json:"file_size"`
	FileSizeText string `json:"file_size_text,omitempty"`
}

var (
	reSignPatterns = []*regexp.Regexp{
		regexp.MustCompile(`'wp_sign'\s*:\s*'([^']+)'`),
		regexp.MustCompile(`"wp_sign"\s*:\s*"([^"]+)"`),
		regexp.MustCompile(`wp_sign\s*=\s*['"]([^'"]+)['"]`),
	}
	reKeyPatterns = []*regexp.Regexp{
		regexp.MustCompile(`'webservicekey'\s*:\s*'([^']+)'`),
		regexp.MustCompile(`"webservicekey"\s*:\s*"([^"]+)"`),
		regexp.MustCompile(`webservicekey\s*=\s*['"]([^'"]+)['"]`),
	}
	reFileNamePatterns = []*regexp.Regexp{
		regexp.MustCompile(`<div[^>]*class="[^"]*filename[^"]*"[^>]*>\s*([^<]+)`),
		regexp.MustCompile(`<title>\s*([^<-]+)`),
	}
	reSizePatterns = []*regexp.Regexp{
		regexp.MustCompile(`<div[^>]*class="[^"]*n_filesize[^"]*"[^>]*>\s*大小[：:]\s*([^<]+)`),
		regexp.MustCompile(`大小[：:]\s*([0-9.]+\s*[KMG]?B?)`),
	}
	reIframePatterns = []*regexp.Regexp{
		regexp.MustCompile(`<iframe[^>]+src="([^"]+)"`),
		regexp.MustCompile(`<iframe[^>]+src='([^']+)'`),
	}
)

const defaultUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// ParseShareLink 解析蓝奏云分享链接，返回直链
func ParseShareLink(shareURL, password string) (*ParseResult, error) {
	shareURL = strings.TrimSpace(shareURL)
	if shareURL == "" {
		return nil, fmt.Errorf("分享链接不能为空")
	}

	parsed, err := url.Parse(shareURL)
	if err != nil {
		return nil, fmt.Errorf("无效的分享链接")
	}
	if !isLanzouDomain(parsed.Host) {
		return nil, fmt.Errorf("不支持的蓝奏云域名: %s", parsed.Host)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("重定向次数过多")
			}
			return nil
		},
	}

	pageURL := shareURL
	html, err := fetchPage(client, pageURL)
	if err != nil {
		return nil, err
	}

	if needsPassword(html) {
		if password == "" {
			return nil, fmt.Errorf("该链接需要访问密码")
		}
		html, pageURL, err = submitPassword(client, pageURL, password)
		if err != nil {
			return nil, err
		}
	}

	if iframeURL := extractFirst(reIframePatterns, html); iframeURL != "" {
		iframeURL, err = resolveURL(pageURL, iframeURL)
		if err != nil {
			return nil, err
		}
		html, err = fetchPage(client, iframeURL)
		if err != nil {
			return nil, err
		}
		pageURL = iframeURL
	}

	sign := extractFirst(reSignPatterns, html)
	key := extractFirst(reKeyPatterns, html)
	if sign == "" || key == "" {
		return nil, fmt.Errorf("解析失败，页面结构可能已变更")
	}

	fileID := extractFileID(pageURL)
	if fileID == "" {
		return nil, fmt.Errorf("无法从链接中提取文件 ID")
	}

	base, _ := url.Parse(pageURL)
	ajaxURL := fmt.Sprintf("%s://%s/ajaxm.php?file=%s", base.Scheme, base.Host, fileID)

	form := url.Values{
		"action":        {"downprocess"},
		"sign":          {sign},
		"webservicekey": {key},
	}
	req, err := http.NewRequest(http.MethodPost, ajaxURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", defaultUA)
	req.Header.Set("Referer", pageURL)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ajaxResp struct {
		Zt  int    `json:"zt"`
		Dom string `json:"dom"`
		URL string `json:"url"`
		Inf string `json:"inf"`
	}
	if err := json.Unmarshal(body, &ajaxResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	if ajaxResp.Zt != 1 {
		msg := ajaxResp.Inf
		if msg == "" {
			msg = "获取直链失败"
		}
		return nil, fmt.Errorf(msg)
	}

	downloadURL := strings.TrimSpace(ajaxResp.Dom + "/file/" + ajaxResp.URL)
	if downloadURL == "" || ajaxResp.URL == "" {
		return nil, fmt.Errorf("获取直链失败")
	}

	fileName := strings.TrimSpace(extractFirst(reFileNamePatterns, html))
	sizeText := strings.TrimSpace(extractFirst(reSizePatterns, html))

	return &ParseResult{
		DownloadURL:  downloadURL,
		FileName:     fileName,
		FileSize:     parseSizeText(sizeText),
		FileSizeText: sizeText,
	}, nil
}

func needsPassword(html string) bool {
	return strings.Contains(html, `id="pass"`) ||
		strings.Contains(html, "输入密码") ||
		strings.Contains(html, "访问密码")
}

func isLanzouDomain(host string) bool {
	host = strings.ToLower(strings.TrimPrefix(host, "www."))
	for _, d := range lanzouDomains {
		if host == d || strings.HasSuffix(host, "."+d) {
			return true
		}
	}
	return strings.Contains(host, "lanzou")
}

func extractFileID(pageURL string) string {
	parsed, err := url.Parse(pageURL)
	if err != nil {
		return ""
	}
	path := strings.Trim(parsed.Path, "/")
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		path = path[idx+1:]
	}
	return path
}

func extractFirst(patterns []*regexp.Regexp, html string) string {
	for _, re := range patterns {
		if m := re.FindStringSubmatch(html); len(m) > 1 {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}

func parseSizeText(text string) int64 {
	text = strings.TrimSpace(strings.ToUpper(text))
	if text == "" {
		return 0
	}
	text = strings.ReplaceAll(text, " ", "")
	multiplier := int64(1)
	switch {
	case strings.HasSuffix(text, "GB"):
		multiplier = 1024 * 1024 * 1024
		text = strings.TrimSuffix(text, "GB")
	case strings.HasSuffix(text, "MB"):
		multiplier = 1024 * 1024
		text = strings.TrimSuffix(text, "MB")
	case strings.HasSuffix(text, "KB"):
		multiplier = 1024
		text = strings.TrimSuffix(text, "KB")
	case strings.HasSuffix(text, "B"):
		text = strings.TrimSuffix(text, "B")
	}
	val, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0
	}
	return int64(val * float64(multiplier))
}

func fetchPage(client *http.Client, pageURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", defaultUA)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func submitPassword(client *http.Client, pageURL, password string) (string, string, error) {
	parsed, err := url.Parse(pageURL)
	if err != nil {
		return "", "", err
	}
	fileID := extractFileID(pageURL)
	postURL := fmt.Sprintf("%s://%s/%s", parsed.Scheme, parsed.Host, fileID)

	form := url.Values{"pwd": {password}}
	req, err := http.NewRequest(http.MethodPost, postURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", defaultUA)
	req.Header.Set("Referer", pageURL)

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	newHTML := string(body)
	if strings.Contains(newHTML, "密码不正确") || needsPassword(newHTML) {
		return "", "", fmt.Errorf("访问密码错误")
	}
	return newHTML, postURL, nil
}

func resolveURL(baseStr, ref string) (string, error) {
	if strings.HasPrefix(ref, "//") {
		ref = "https:" + ref
	}
	base, err := url.Parse(baseStr)
	if err != nil {
		return "", err
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(refURL).String(), nil
}
