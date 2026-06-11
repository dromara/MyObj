package util

import "strings"

// IsDocumentMime 判断是否为文档类型
func IsDocumentMime(mime string) bool {
	return strings.HasPrefix(mime, "application/pdf") ||
		strings.Contains(mime, "word") ||
		strings.Contains(mime, "excel") ||
		strings.Contains(mime, "spreadsheet") ||
		strings.Contains(mime, "powerpoint") ||
		strings.Contains(mime, "presentation") ||
		strings.HasPrefix(mime, "text/")
}

// IsArchiveMime 判断是否为压缩包类型
func IsArchiveMime(mime string) bool {
	return strings.Contains(mime, "zip") ||
		strings.Contains(mime, "rar") ||
		strings.Contains(mime, "tar") ||
		strings.Contains(mime, "gzip") ||
		strings.Contains(mime, "7z") ||
		strings.Contains(mime, "compress")
}
