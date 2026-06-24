package util

import (
	"path/filepath"
	"strings"
)

var extCategoryMap = map[string]string{
	".doc": "document", ".docx": "document", ".pdf": "document", ".xls": "document",
	".xlsx": "document", ".ppt": "document", ".pptx": "document", ".txt": "document",
	".md": "document", ".rtf": "document", ".csv": "document",
	".jpg": "image", ".jpeg": "image", ".png": "image", ".gif": "image",
	".bmp": "image", ".webp": "image", ".svg": "image", ".ico": "image",
	".mp4": "video", ".avi": "video", ".mkv": "video", ".mov": "video",
	".wmv": "video", ".flv": "video", ".webm": "video", ".m4v": "video",
	".mp3": "audio", ".wav": "audio", ".flac": "audio", ".aac": "audio",
	".ogg": "audio", ".wma": "audio", ".m4a": "audio",
	".zip": "archive", ".rar": "archive", ".7z": "archive", ".tar": "archive",
	".gz": "archive", ".bz2": "archive", ".xz": "archive",
	".go": "code", ".js": "code", ".ts": "code", ".py": "code",
	".java": "code", ".c": "code", ".cpp": "code", ".html": "code",
	".css": "code", ".json": "code", ".xml": "code", ".yaml": "code",
	".sql": "code", ".sh": "code", ".rb": "code", ".php": "code",
}

func ClassifyByExt(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	if cat, ok := extCategoryMap[ext]; ok {
		return cat
	}
	return "other"
}

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
