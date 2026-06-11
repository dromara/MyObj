package extract

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bodgit/sevenzip"
	"github.com/klauspost/compress/zstd"
	"github.com/nwaples/rardecode/v2"
	"github.com/ulikunitz/xz"
	"myobj/src/pkg/logger"
)

// ArchiveType 归档类型
type ArchiveType int

const (
	ArchiveTypeZIP   ArchiveType = iota
	ArchiveTypeTAR
	ArchiveTypeTARGZ
	ArchiveTypeTARBZ2
	ArchiveTypeTARXZ
	ArchiveTypeTARZST
	ArchiveType7Z
	ArchiveTypeRAR
	ArchiveTypeUnknown
)

// ExtractProgress 提取进度
type ExtractProgress struct {
	CurrentFile  string `json:"current_file"`
	CurrentIndex int    `json:"current_index"`
	TotalFiles   int    `json:"total_files"`
	CurrentSize  int64  `json:"current_size"`
	TotalSize    int64  `json:"total_size"`
}

// ExtractedEntry 提取出的条目信息
type ExtractedEntry struct {
	FilePath string
	FileName string
	FileSize int64
	IsDir    bool
	MimeType string
}

// ExtractResult 提取结果
type ExtractResult struct {
	Entries     []ExtractedEntry
	TotalFiles  int
	TotalSize   int64
	ArchiveName string
	ArchiveType string
}

// ExtractOptions 提取选项
type ExtractOptions struct {
	ProgressCallback func(progress ExtractProgress)
	MaxFileSize      int64
	MaxTotalSize     int64
}

func DetectArchiveType(filename string) ArchiveType {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return ArchiveTypeZIP
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return ArchiveTypeTARGZ
	case strings.HasSuffix(lower, ".tar.bz2"), strings.HasSuffix(lower, ".tbz2"), strings.HasSuffix(lower, ".tbz"):
		return ArchiveTypeTARBZ2
	case strings.HasSuffix(lower, ".tar.xz"), strings.HasSuffix(lower, ".txz"):
		return ArchiveTypeTARXZ
	case strings.HasSuffix(lower, ".tar.zst"), strings.HasSuffix(lower, ".tzst"):
		return ArchiveTypeTARZST
	case strings.HasSuffix(lower, ".tar"):
		return ArchiveTypeTAR
	case strings.HasSuffix(lower, ".7z"):
		return ArchiveType7Z
	case strings.HasSuffix(lower, ".rar"):
		return ArchiveTypeRAR
	default:
		return ArchiveTypeUnknown
	}
}

func IsSupportedArchive(filename string) bool {
	return DetectArchiveType(filename) != ArchiveTypeUnknown
}

func GetSupportedFormats() []string {
	return []string{".zip", ".tar", ".tar.gz", ".tgz", ".tar.bz2", ".tbz2", ".tbz", ".tar.xz", ".txz", ".tar.zst", ".tzst", ".7z", ".rar"}
}

func SanitizePath(baseDir, entryPath string) (string, error) {
	cleanPath := filepath.Clean(entryPath)
	if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, "\\..") {
		return "", fmt.Errorf("path traversal detected: %s", entryPath)
	}
	fullPath := filepath.Join(baseDir, cleanPath)
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve base dir: %w", err)
	}
	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("cannot resolve target path: %w", err)
	}
	if !strings.HasPrefix(absFull, absBase+string(filepath.Separator)) && absFull != absBase {
		return "", fmt.Errorf("path escape detected: %s", entryPath)
	}
	return absFull, nil
}

func ExtractArchive(archivePath, destDir string, opts *ExtractOptions) (*ExtractResult, error) {
	archiveType := DetectArchiveType(archivePath)
	switch archiveType {
	case ArchiveTypeZIP:
		return extractZip(archivePath, destDir, opts)
	case ArchiveTypeTAR:
		return extractTar(archivePath, destDir, opts, nil)
	case ArchiveTypeTARGZ:
		return extractTarGz(archivePath, destDir, opts)
	case ArchiveTypeTARBZ2:
		return extractTarBz2(archivePath, destDir, opts)
	case ArchiveType7Z:
		return extract7z(archivePath, destDir, opts)
	case ArchiveTypeRAR:
		return extractRar(archivePath, destDir, opts)
	case ArchiveTypeTARXZ:
		return extractTarXz(archivePath, destDir, opts)
	case ArchiveTypeTARZST:
		return extractTarZst(archivePath, destDir, opts)
	default:
		return nil, fmt.Errorf("unsupported archive format: %s", filepath.Ext(archivePath))
	}
}

// ListArchiveEntries 列出压缩包中的文件名（不解压内容）
// 返回所有非目录条目的文件名列表
func ListArchiveEntries(archivePath string) ([]string, error) {
	archiveType := DetectArchiveType(archivePath)
	switch archiveType {
	case ArchiveTypeZIP:
		return listZipEntries(archivePath)
	case ArchiveTypeTAR:
		return listTarEntries(archivePath, nil)
	case ArchiveTypeTARGZ:
		return listTarEntries(archivePath, func(r io.Reader) (io.ReadCloser, error) {
			return gzip.NewReader(r)
		})
	case ArchiveTypeTARBZ2:
		return listTarEntries(archivePath, func(r io.Reader) (io.ReadCloser, error) {
			return io.NopCloser(bzip2.NewReader(r)), nil
		})
	case ArchiveTypeTARXZ:
		return listTarEntries(archivePath, func(r io.Reader) (io.ReadCloser, error) {
			xzReader, err := xz.NewReader(r)
			if err != nil {
				return nil, err
			}
			return io.NopCloser(xzReader), nil
		})
	case ArchiveTypeTARZST:
		return listTarEntries(archivePath, func(r io.Reader) (io.ReadCloser, error) {
			decoder, err := zstd.NewReader(r)
			if err != nil {
				return nil, err
			}
			return decoder.IOReadCloser(), nil
		})
	case ArchiveType7Z:
		return list7zEntries(archivePath)
	case ArchiveTypeRAR:
		return listRarEntries(archivePath)
	default:
		return nil, fmt.Errorf("unsupported archive format: %s", filepath.Ext(archivePath))
	}
}

func listZipEntries(zipPath string) ([]string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP: %w", err)
	}
	defer reader.Close()

	var names []string
	for _, f := range reader.File {
		if !f.FileInfo().IsDir() {
			names = append(names, filepath.Base(f.Name))
		}
	}
	return names, nil
}

func listTarEntries(tarPath string, decompressor func(io.Reader) (io.ReadCloser, error)) ([]string, error) {
	file, err := os.Open(tarPath)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file
	if decompressor != nil {
		rc, err := decompressor(file)
		if err != nil {
			return nil, fmt.Errorf("create decompressor failed: %w", err)
		}
		defer rc.Close()
		reader = rc
	}

	tarReader := tar.NewReader(reader)
	var names []string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar entry failed: %w", err)
		}
		if header.Typeflag == tar.TypeReg {
			names = append(names, filepath.Base(header.Name))
		}
	}
	return names, nil
}

func list7zEntries(sevenzPath string) ([]string, error) {
	reader, err := sevenzip.OpenReader(sevenzPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open 7z: %w", err)
	}
	defer reader.Close()

	var names []string
	for _, f := range reader.File {
		if !f.FileInfo().IsDir() {
			names = append(names, filepath.Base(f.Name))
		}
	}
	return names, nil
}

func listRarEntries(rarPath string) ([]string, error) {
	rc, err := rardecode.OpenReader(rarPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open RAR: %w", err)
	}
	defer rc.Close()

	var names []string
	for {
		hdr, err := rc.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read RAR entry failed: %w", err)
		}
		if !hdr.IsDir {
			names = append(names, filepath.Base(hdr.Name))
		}
	}
	return names, nil
}

func extractZip(zipPath, destDir string, opts *ExtractOptions) (*ExtractResult, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP: %w", err)
	}
	defer reader.Close()

	var (
		entries   []ExtractedEntry
		mu        sync.Mutex
		totalSize int64
		fileCount int
	)

	for _, f := range reader.File {
		if !f.FileInfo().IsDir() {
			fileCount++
			totalSize += int64(f.UncompressedSize64)
		}
	}

	if opts != nil && opts.MaxTotalSize > 0 && totalSize > opts.MaxTotalSize {
		return nil, fmt.Errorf("extract total size(%d) exceeds limit(%d)", totalSize, opts.MaxTotalSize)
	}

	currentIndex := 0
	for _, f := range reader.File {
		safePath, err := SanitizePath(destDir, f.Name)
		if err != nil {
			logger.LOG.Warn("skip unsafe path", "path", f.Name, "error", err)
			continue
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(safePath, 0755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(safePath), 0755); err != nil {
			return nil, fmt.Errorf("create parent dir failed: %w", err)
		}
		if opts != nil && opts.MaxFileSize > 0 && int64(f.UncompressedSize64) > opts.MaxFileSize {
			logger.LOG.Warn("skip large file", "name", f.Name, "size", f.UncompressedSize64)
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open entry failed [%s]: %w", f.Name, err)
		}
		outFile, err := os.Create(safePath)
		if err != nil {
			rc.Close()
			return nil, fmt.Errorf("create file failed [%s]: %w", f.Name, err)
		}
		// 使用 LimitReader 限制实际读取字节数，防止 ZIP bomb
		maxSize := int64(f.UncompressedSize64)
		if maxSize <= 0 {
			maxSize = 1 << 30 // 默认上限 1GB
		}
		limitedReader := io.LimitReader(rc, maxSize+1) // +1 用于检测是否超出
		written, err := io.Copy(outFile, limitedReader)
		rc.Close()
		outFile.Close()
		if err != nil {
			return nil, fmt.Errorf("write file failed [%s]: %w", f.Name, err)
		}
		if written > maxSize {
			return nil, fmt.Errorf("ZIP bomb detected: file [%s] exceeds declared uncompressed size %d", f.Name, maxSize)
		}
		currentIndex++
		entry := ExtractedEntry{
			FilePath: safePath,
			FileName: filepath.Base(f.Name),
			FileSize: int64(f.UncompressedSize64),
			IsDir:    false,
		}
		mu.Lock()
		entries = append(entries, entry)
		mu.Unlock()
		if opts != nil && opts.ProgressCallback != nil {
			opts.ProgressCallback(ExtractProgress{
				CurrentFile:  f.Name,
				CurrentIndex: currentIndex,
				TotalFiles:   fileCount,
				TotalSize:    totalSize,
			})
		}
	}

	return &ExtractResult{
		Entries:     entries,
		TotalFiles:  fileCount,
		TotalSize:   totalSize,
		ArchiveName: filepath.Base(zipPath),
		ArchiveType: "zip",
	}, nil
}

func extractTar(tarPath, destDir string, opts *ExtractOptions, decompressor func(io.Reader) (io.ReadCloser, error)) (*ExtractResult, error) {
	// Delegate to single-pass
	archiveType := "tar"
	if strings.HasSuffix(strings.ToLower(tarPath), ".gz") || strings.HasSuffix(strings.ToLower(tarPath), ".tgz") {
		archiveType = "tar.gz"
	}
	if strings.HasSuffix(strings.ToLower(tarPath), ".bz2") || strings.HasSuffix(strings.ToLower(tarPath), ".tbz2") || strings.HasSuffix(strings.ToLower(tarPath), ".tbz") {
		archiveType = "tar.bz2"
	}
	if strings.HasSuffix(strings.ToLower(tarPath), ".xz") || strings.HasSuffix(strings.ToLower(tarPath), ".txz") {
		archiveType = "tar.xz"
	}
	if strings.HasSuffix(strings.ToLower(tarPath), ".zst") || strings.HasSuffix(strings.ToLower(tarPath), ".tzst") {
		archiveType = "tar.zst"
	}
	return extractTarSinglePass(tarPath, destDir, opts, decompressor, archiveType)
}

func extractTarGz(tarGzPath, destDir string, opts *ExtractOptions) (*ExtractResult, error) {
	return extractTar(tarGzPath, destDir, opts, func(r io.Reader) (io.ReadCloser, error) {
		return gzip.NewReader(r)
	})
}

func extract7z(sevenzPath, destDir string, opts *ExtractOptions) (*ExtractResult, error) {
	reader, err := sevenzip.OpenReader(sevenzPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open 7z: %w", err)
	}
	defer reader.Close()

	var (
		entries   []ExtractedEntry
		mu        sync.Mutex
		totalSize int64
		fileCount int
	)

	for _, f := range reader.File {
		if !f.FileInfo().IsDir() {
			fileCount++
			totalSize += f.FileInfo().Size()
		}
	}

	if opts != nil && opts.MaxTotalSize > 0 && totalSize > opts.MaxTotalSize {
		return nil, fmt.Errorf("extract total size(%d) exceeds limit(%d)", totalSize, opts.MaxTotalSize)
	}

	currentIndex := 0
	for _, f := range reader.File {
		safePath, err := SanitizePath(destDir, f.Name)
		if err != nil {
			logger.LOG.Warn("skip unsafe path", "path", f.Name, "error", err)
			continue
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(safePath, 0755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(safePath), 0755); err != nil {
			return nil, fmt.Errorf("create parent dir failed: %w", err)
		}
		fileSize := f.FileInfo().Size()
		if opts != nil && opts.MaxFileSize > 0 && fileSize > opts.MaxFileSize {
			logger.LOG.Warn("skip large file", "name", f.Name, "size", fileSize)
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open entry failed [%s]: %w", f.Name, err)
		}
		outFile, err := os.Create(safePath)
		if err != nil {
			rc.Close()
			return nil, fmt.Errorf("create file failed [%s]: %w", f.Name, err)
		}
		// 使用 LimitReader 限制实际读取字节数，防止 ZIP bomb
		maxSize := fileSize
		if maxSize <= 0 {
			maxSize = 1 << 30 // 默认上限 1GB
		}
		limitedReader := io.LimitReader(rc, maxSize+1)
		written, err := io.Copy(outFile, limitedReader)
		rc.Close()
		outFile.Close()
		if err != nil {
			return nil, fmt.Errorf("write file failed [%s]: %w", f.Name, err)
		}
		if written > maxSize {
			return nil, fmt.Errorf("ZIP bomb detected: file [%s] exceeds declared uncompressed size %d", f.Name, maxSize)
		}
		currentIndex++
		entry := ExtractedEntry{
			FilePath: safePath,
			FileName: filepath.Base(f.Name),
			FileSize: fileSize,
			IsDir:    false,
		}
		mu.Lock()
		entries = append(entries, entry)
		mu.Unlock()
		if opts != nil && opts.ProgressCallback != nil {
			opts.ProgressCallback(ExtractProgress{
				CurrentFile:  f.Name,
				CurrentIndex: currentIndex,
				TotalFiles:   fileCount,
				TotalSize:    totalSize,
			})
		}
	}

	return &ExtractResult{
		Entries:     entries,
		TotalFiles:  fileCount,
		TotalSize:   totalSize,
		ArchiveName: filepath.Base(sevenzPath),
		ArchiveType: "7z",
	}, nil
}

func extractTarBz2(tarBz2Path, destDir string, opts *ExtractOptions) (*ExtractResult, error) {
	return extractTar(tarBz2Path, destDir, opts, func(r io.Reader) (io.ReadCloser, error) {
		return io.NopCloser(bzip2.NewReader(r)), nil
	})
}

func extractTarXz(tarXzPath, destDir string, opts *ExtractOptions) (*ExtractResult, error) {
	return extractTar(tarXzPath, destDir, opts, func(r io.Reader) (io.ReadCloser, error) {
		xzReader, err := xz.NewReader(r)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(xzReader), nil
	})
}

func extractTarZst(tarZstPath, destDir string, opts *ExtractOptions) (*ExtractResult, error) {
	return extractTar(tarZstPath, destDir, opts, func(r io.Reader) (io.ReadCloser, error) {
		decoder, err := zstd.NewReader(r)
		if err != nil {
			return nil, err
		}
		return decoder.IOReadCloser(), nil
	})
}

func extractRar(rarPath, destDir string, opts *ExtractOptions) (*ExtractResult, error) {
	rc, err := rardecode.OpenReader(rarPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open RAR: %w", err)
	}
	defer rc.Close()

	var (
		entries   []ExtractedEntry
		totalSize int64
		fileCount int
	)

	// First pass: count files and total size
	for {
		hdr, err := rc.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read RAR entry failed: %w", err)
		}
		if !hdr.IsDir {
			fileCount++
			totalSize += hdr.UnPackedSize
		}
	}

	if opts != nil && opts.MaxTotalSize > 0 && totalSize > opts.MaxTotalSize {
		return nil, fmt.Errorf("extract total size(%d) exceeds limit(%d)", totalSize, opts.MaxTotalSize)
	}

	// Reopen for second pass
	if err := rc.Close(); err != nil {
		return nil, fmt.Errorf("close RAR for reopen failed: %w", err)
	}
	rc, err = rardecode.OpenReader(rarPath)
	if err != nil {
		return nil, fmt.Errorf("failed to reopen RAR: %w", err)
	}
	defer rc.Close()

	currentIndex := 0
	for {
		hdr, err := rc.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read RAR entry failed: %w", err)
		}

		safePath, err := SanitizePath(destDir, hdr.Name)
		if err != nil {
			logger.LOG.Warn("skip unsafe path", "path", hdr.Name, "error", err)
			continue
		}
		if hdr.IsDir {
			os.MkdirAll(safePath, 0755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(safePath), 0755); err != nil {
			return nil, fmt.Errorf("create parent dir failed: %w", err)
		}
		fileSize := hdr.UnPackedSize
		if opts != nil && opts.MaxFileSize > 0 && fileSize > opts.MaxFileSize {
			logger.LOG.Warn("skip large file", "name", hdr.Name, "size", fileSize)
			continue
		}
		outFile, err := os.Create(safePath)
		if err != nil {
			return nil, fmt.Errorf("create file failed [%s]: %w", hdr.Name, err)
		}
		// 使用 LimitReader 限制实际读取字节数，防止解压炸弹
		maxSize := fileSize
		if maxSize <= 0 {
			maxSize = 1 << 30 // 默认上限 1GB
		}
		limitedReader := io.LimitReader(rc, maxSize+1) // +1 用于检测是否超出
		written, err := io.Copy(outFile, limitedReader)
		outFile.Close()
		if err != nil {
			os.Remove(safePath)
			return nil, fmt.Errorf("write file failed [%s]: %w", hdr.Name, err)
		}
		if written > maxSize {
			os.Remove(safePath)
			return nil, fmt.Errorf("RAR entry exceeds max file size (possible decompression bomb): %s", hdr.Name)
		}
		currentIndex++
		entry := ExtractedEntry{
			FilePath: safePath,
			FileName: filepath.Base(hdr.Name),
			FileSize: fileSize,
			IsDir:    false,
		}
		entries = append(entries, entry)
		if opts != nil && opts.ProgressCallback != nil {
			opts.ProgressCallback(ExtractProgress{
				CurrentFile:  hdr.Name,
				CurrentIndex: currentIndex,
				TotalFiles:   fileCount,
				TotalSize:    totalSize,
			})
		}
	}

	return &ExtractResult{
		Entries:     entries,
		TotalFiles:  fileCount,
		TotalSize:   totalSize,
		ArchiveName: filepath.Base(rarPath),
		ArchiveType: "rar",
	}, nil
}

func extractTarSinglePass(tarPath, destDir string, opts *ExtractOptions, decompressor func(io.Reader) (io.ReadCloser, error), archiveType string) (*ExtractResult, error) {
	file, err := os.Open(tarPath)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file
	if decompressor != nil {
		rc, err := decompressor(file)
		if err != nil {
			return nil, fmt.Errorf("create decompressor failed: %w", err)
		}
		defer rc.Close()
		reader = rc
	}

	tarReader := tar.NewReader(reader)
	var (
		entries   []ExtractedEntry
		totalSize int64
		fileCount int
	)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar entry failed: %w", err)
		}
		safePath, err := SanitizePath(destDir, header.Name)
		if err != nil {
			logger.LOG.Warn("skip unsafe path", "path", header.Name, "error", err)
			continue
		}
		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(safePath, 0755)
		case tar.TypeReg:
			if opts != nil && opts.MaxFileSize > 0 && header.Size > opts.MaxFileSize {
				logger.LOG.Warn("skip large file", "name", header.Name, "size", header.Size)
				continue
			}
			totalSize += header.Size
			if opts != nil && opts.MaxTotalSize > 0 && totalSize > opts.MaxTotalSize {
				return nil, fmt.Errorf("extract total size exceeds limit(%d)", opts.MaxTotalSize)
			}
			if err := os.MkdirAll(filepath.Dir(safePath), 0755); err != nil {
				return nil, fmt.Errorf("create parent dir failed: %w", err)
			}
			outFile, err := os.Create(safePath)
			if err != nil {
				return nil, fmt.Errorf("create file failed [%s]: %w", header.Name, err)
			}
			// 使用 LimitReader 限制实际读取字节数，防止解压炸弹
			maxSize := header.Size
			if maxSize <= 0 {
				maxSize = 1 << 30 // 默认上限 1GB
			}
			limitedReader := io.LimitReader(tarReader, maxSize+1) // +1 用于检测是否超出
			written, err := io.Copy(outFile, limitedReader)
			outFile.Close()
			if err != nil {
				os.Remove(safePath)
				return nil, fmt.Errorf("write file failed [%s]: %w", header.Name, err)
			}
			if written > maxSize {
				os.Remove(safePath)
				return nil, fmt.Errorf("TAR entry exceeds max file size (possible decompression bomb): %s", header.Name)
			}
			fileCount++
			entry := ExtractedEntry{
				FilePath: safePath,
				FileName: filepath.Base(header.Name),
				FileSize: header.Size,
				IsDir:    false,
			}
			entries = append(entries, entry)
			if opts != nil && opts.ProgressCallback != nil {
				opts.ProgressCallback(ExtractProgress{
					CurrentFile:  header.Name,
					CurrentIndex: fileCount,
					TotalFiles:   fileCount,
					CurrentSize:  totalSize,
					TotalSize:    totalSize,
				})
			}
		}
	}

	return &ExtractResult{
		Entries:     entries,
		TotalFiles:  fileCount,
		TotalSize:   totalSize,
		ArchiveName: filepath.Base(tarPath),
		ArchiveType: archiveType,
	}, nil
}
