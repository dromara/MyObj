package util

import (
	"fmt"
	"net/url"
	"strings"
)

// BuildContentDisposition 构建安全的 Content-Disposition header
// disposition: "attachment" 或 "inline"
func BuildContentDisposition(fileName string, disposition string) string {
	escapedName := strings.ReplaceAll(fileName, "\"", "\\\"")
	encodedName := url.PathEscape(fileName)
	return fmt.Sprintf(`%s; filename="%s"; filename*=UTF-8''%s`, disposition, escapedName, encodedName)
}
