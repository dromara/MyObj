package internal

import (
	"net/http"
	"strings"
)

// SetCookieValue 更新 cookie 字符串中指定 key 的值
func SetCookieValue(cookieStr, name, value string) string {
	cookies := parseCookies(cookieStr)
	found := false
	for i, c := range cookies {
		if c.Name == name {
			cookies[i].Value = value
			found = true
			break
		}
	}
	if !found {
		cookies = append(cookies, &http.Cookie{Name: name, Value: value})
	}
	return cookiesToString(cookies)
}

func parseCookies(s string) []*http.Cookie {
	header := http.Header{}
	header.Add("Cookie", s)
	req := &http.Request{Header: header}
	return req.Cookies()
}

func cookiesToString(cookies []*http.Cookie) string {
	parts := make([]string, len(cookies))
	for i, c := range cookies {
		parts[i] = c.Name + "=" + c.Value
	}
	return strings.Join(parts, "; ")
}
