package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"myobj/src/core/domain/response"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/models"
	"sort"
	"strings"
	"unicode"

	"gorm.io/gorm"
)

type CinemaService struct {
	factory    *impl.RepositoryFactory
	tagService *TagService
}

func NewCinemaService(factory *impl.RepositoryFactory, tagService *TagService) *CinemaService {
	return &CinemaService{factory: factory, tagService: tagService}
}

type cinemaVideoRow struct {
	FileID       string
	FileName     string
	DirectoryID  int
	FileSize     int
	MimeType     string
	IsEnc        bool
	Public       bool
	ThumbnailImg string
	CreatedAt    custom_type.JsonTime
}

type cinemaTree struct {
	root        models.VirtualDirectory
	directories []models.VirtualDirectory
	byID        map[int]models.VirtualDirectory
	paths       map[int]string
	ids         []int
}

func normalizeCinemaPage(page, pageSize, defaultSize, maxSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > maxSize {
		pageSize = defaultSize
	}
	return page, pageSize
}

func (s *CinemaService) loadTree(ctx context.Context, userID string, rootID int) (*cinemaTree, error) {
	var root models.VirtualDirectory
	if err := s.factory.DB().WithContext(ctx).Where("id = ? AND user_id = ?", rootID, userID).First(&root).Error; err != nil {
		return nil, errors.New("影视文件夹不存在或无权访问")
	}
	enabled, err := s.tagService.IsCinemaDirectory(ctx, userID, rootID)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, errors.New("该文件夹未启用影视模式")
	}
	var all []models.VirtualDirectory
	if err := s.factory.DB().WithContext(ctx).Where("user_id = ?", userID).Find(&all).Error; err != nil {
		return nil, err
	}
	children := make(map[int][]models.VirtualDirectory)
	for _, directory := range all {
		children[directory.ParentID] = append(children[directory.ParentID], directory)
	}
	result := &cinemaTree{
		root: root, directories: []models.VirtualDirectory{root},
		byID: map[int]models.VirtualDirectory{root.ID: root}, paths: map[int]string{root.ID: root.Name}, ids: []int{root.ID},
	}
	for index := 0; index < len(result.directories); index++ {
		parent := result.directories[index]
		for _, child := range children[parent.ID] {
			result.directories = append(result.directories, child)
			result.byID[child.ID] = child
			result.paths[child.ID] = result.paths[parent.ID] + "/" + child.Name
			result.ids = append(result.ids, child.ID)
		}
	}
	return result, nil
}

func cinemaDirectoryView(tree *cinemaTree, directory models.VirtualDirectory) response.CinemaDirectory {
	return response.CinemaDirectory{ID: directory.ID, Name: directory.Name, ParentID: directory.ParentID, Path: tree.paths[directory.ID]}
}

func (s *CinemaService) playableVideoQuery(ctx context.Context, userID string, directoryIDs []int) *gorm.DB {
	return s.factory.DB().WithContext(ctx).Table("user_files AS uf").
		Select(`uf.uf_id AS file_id, uf.file_name, uf.directory_id, fi.size AS file_size,
			fi.mime AS mime_type, fi.is_enc, uf.public, fi.thumbnail_img, uf.created_at`).
		Joins("JOIN file_info fi ON fi.id = uf.file_id").
		Where("uf.user_id = ? AND uf.directory_id IN ? AND uf.deleted_at IS NULL", userID, directoryIDs).
		Where("fi.mime LIKE ? AND (fi.path <> '' OR fi.enc_path <> '')", "video/%")
}

func (s *CinemaService) loadVideoRows(ctx context.Context, userID string, directoryIDs []int, offset, limit int) ([]cinemaVideoRow, int64, error) {
	query := s.playableVideoQuery(ctx, userID, directoryIDs)
	var total int64
	if err := query.Select("count(*)").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []cinemaVideoRow
	if err := s.playableVideoQuery(ctx, userID, directoryIDs).
		Order("uf.created_at DESC, uf.uf_id DESC").Offset(offset).Limit(limit).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (s *CinemaService) buildVideoItems(ctx context.Context, userID string, tree *cinemaTree, rows []cinemaVideoRow) ([]response.CinemaVideoItem, error) {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.FileID)
	}
	tags, err := s.tagService.CompactTags(ctx, userID, userID, ids, false)
	if err != nil {
		return nil, err
	}
	items := make([]response.CinemaVideoItem, 0, len(rows))
	for _, row := range rows {
		directory, ok := tree.byID[row.DirectoryID]
		if !ok {
			continue
		}
		items = append(items, response.CinemaVideoItem{
			FileID: row.FileID, FileName: row.FileName, FileSize: row.FileSize, MimeType: row.MimeType,
			IsEnc: row.IsEnc, Public: row.Public, HasThumbnail: row.ThumbnailImg != "", CreatedAt: row.CreatedAt,
			Directory: cinemaDirectoryView(tree, directory), Tags: tags[row.FileID],
		})
	}
	return items, nil
}

func (s *CinemaService) Home(ctx context.Context, userID string, rootID, page, pageSize int) (*response.CinemaHomeResponse, error) {
	tree, err := s.loadTree(ctx, userID, rootID)
	if err != nil {
		return nil, err
	}
	type countRow struct {
		DirectoryID int
		Total       int64
	}
	var counts []countRow
	if err := s.playableVideoQuery(ctx, userID, tree.ids).
		Select("uf.directory_id, count(*) AS total").Group("uf.directory_id").Scan(&counts).Error; err != nil {
		return nil, err
	}
	countByID := make(map[int]int64, len(counts))
	for _, count := range counts {
		countByID[count.DirectoryID] = count.Total
	}
	descendants := make([]models.VirtualDirectory, 0, len(counts))
	for _, directory := range tree.directories {
		if directory.ID != rootID && countByID[directory.ID] > 0 {
			descendants = append(descendants, directory)
		}
	}
	sort.Slice(descendants, func(i, j int) bool {
		left := strings.ToLower(tree.paths[descendants[i].ID])
		right := strings.ToLower(tree.paths[descendants[j].ID])
		if left != right {
			return left < right
		}
		if tree.paths[descendants[i].ID] != tree.paths[descendants[j].ID] {
			return tree.paths[descendants[i].ID] < tree.paths[descendants[j].ID]
		}
		return descendants[i].ID < descendants[j].ID
	})
	directories := make([]models.VirtualDirectory, 0, len(descendants)+1)
	if countByID[rootID] > 0 {
		directories = append(directories, tree.root)
	}
	directories = append(directories, descendants...)
	page, pageSize = normalizeCinemaPage(page, pageSize, 20, 50)
	start := (page - 1) * pageSize
	if start > len(directories) {
		start = len(directories)
	}
	end := start + pageSize
	if end > len(directories) {
		end = len(directories)
	}
	sections := make([]response.CinemaSection, 0, end-start)
	for _, directory := range directories[start:end] {
		rows, _, err := s.loadVideoRows(ctx, userID, []int{directory.ID}, 0, 6)
		if err != nil {
			return nil, err
		}
		videos, err := s.buildVideoItems(ctx, userID, tree, rows)
		if err != nil {
			return nil, err
		}
		sections = append(sections, response.CinemaSection{
			Directory: cinemaDirectoryView(tree, directory), Videos: videos,
			Total: countByID[directory.ID], HasMore: countByID[directory.ID] > int64(len(videos)),
		})
	}
	return &response.CinemaHomeResponse{
		Root: cinemaDirectoryView(tree, tree.root), Sections: sections, Total: len(directories),
		Page: page, PageSize: pageSize, HasMore: end < len(directories),
	}, nil
}

func (s *CinemaService) FolderVideos(ctx context.Context, userID string, rootID, directoryID, page, pageSize int) (*response.CinemaVideoListResponse, error) {
	tree, err := s.loadTree(ctx, userID, rootID)
	if err != nil {
		return nil, err
	}
	directory, ok := tree.byID[directoryID]
	if !ok {
		return nil, errors.New("文件夹不在当前影视目录中")
	}
	page, pageSize = normalizeCinemaPage(page, pageSize, 24, 100)
	rows, total, err := s.loadVideoRows(ctx, userID, []int{directoryID}, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	videos, err := s.buildVideoItems(ctx, userID, tree, rows)
	if err != nil {
		return nil, err
	}
	return &response.CinemaVideoListResponse{
		Root: cinemaDirectoryView(tree, tree.root), Directory: cinemaDirectoryView(tree, directory), Videos: videos,
		Total: total, Page: page, PageSize: pageSize, HasMore: int64(page*pageSize) < total,
	}, nil
}

func (s *CinemaService) Latest(ctx context.Context, userID string, rootID, page, pageSize int) (*response.CinemaLatestResponse, error) {
	tree, err := s.loadTree(ctx, userID, rootID)
	if err != nil {
		return nil, err
	}
	page, pageSize = normalizeCinemaPage(page, pageSize, 24, 100)
	rows, total, err := s.loadVideoRows(ctx, userID, tree.ids, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	videos, err := s.buildVideoItems(ctx, userID, tree, rows)
	if err != nil {
		return nil, err
	}
	return &response.CinemaLatestResponse{
		Videos: videos, Total: total, Page: page, PageSize: pageSize, HasMore: int64(page*pageSize) < total,
	}, nil
}

func (s *CinemaService) VideoDetail(ctx context.Context, userID string, rootID int, fileID string) (*response.CinemaVideoDetailResponse, error) {
	tree, err := s.loadTree(ctx, userID, rootID)
	if err != nil {
		return nil, err
	}
	var rows []cinemaVideoRow
	if err := s.playableVideoQuery(ctx, userID, tree.ids).Where("uf.uf_id = ?", fileID).Limit(1).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, errors.New("视频不存在或不在当前影视目录中")
	}
	items, err := s.buildVideoItems(ctx, userID, tree, rows)
	if err != nil {
		return nil, err
	}
	return &response.CinemaVideoDetailResponse{Root: cinemaDirectoryView(tree, tree.root), Video: items[0]}, nil
}

func filenameTokens(name string) map[string]struct{} {
	if index := strings.LastIndex(name, "."); index > 0 {
		name = name[:index]
	}
	parts := strings.FieldsFunc(strings.ToLower(name), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })
	result := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		runes := []rune(part)
		if len(runes) > 1 {
			result[part] = struct{}{}
		}
		hasHan := false
		for _, value := range runes {
			if unicode.Is(unicode.Han, value) {
				hasHan = true
				break
			}
		}
		if hasHan {
			for index := 0; index+1 < len(runes); index++ {
				result[string(runes[index:index+2])] = struct{}{}
			}
		}
	}
	return result
}

func overlapCount[T comparable](left, right map[T]struct{}) int {
	count := 0
	for value := range left {
		if _, ok := right[value]; ok {
			count++
		}
	}
	return count
}

func (s *CinemaService) Related(ctx context.Context, userID string, rootID int, fileID string, page, pageSize int) (*response.CinemaRelatedResponse, error) {
	tree, err := s.loadTree(ctx, userID, rootID)
	if err != nil {
		return nil, err
	}
	rows, _, err := s.loadVideoRows(ctx, userID, tree.ids, 0, -1)
	if err != nil {
		return nil, err
	}
	var current *cinemaVideoRow
	for index := range rows {
		if rows[index].FileID == fileID {
			current = &rows[index]
			break
		}
	}
	if current == nil {
		return nil, errors.New("视频不存在或不在当前影视目录中")
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.FileID)
	}
	tags, err := s.tagService.CompactTags(ctx, userID, userID, ids, false)
	if err != nil {
		return nil, err
	}
	currentTags := make(map[string]struct{})
	for _, tag := range tags[fileID] {
		currentTags[tag.ID] = struct{}{}
	}
	currentTokens := filenameTokens(current.FileName)
	type rankedVideo struct {
		row          cinemaVideoRow
		sharedTags   int
		sharedTokens int
	}
	ranked := make([]rankedVideo, 0, len(rows)-1)
	for _, row := range rows {
		if row.FileID == fileID {
			continue
		}
		rowTags := make(map[string]struct{})
		for _, tag := range tags[row.FileID] {
			rowTags[tag.ID] = struct{}{}
		}
		ranked = append(ranked, rankedVideo{
			row: row, sharedTags: overlapCount(currentTags, rowTags), sharedTokens: overlapCount(currentTokens, filenameTokens(row.FileName)),
		})
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].sharedTags != ranked[j].sharedTags {
			return ranked[i].sharedTags > ranked[j].sharedTags
		}
		if ranked[i].sharedTokens != ranked[j].sharedTokens {
			return ranked[i].sharedTokens > ranked[j].sharedTokens
		}
		left, right := ranked[i].row.CreatedAt.ToTime(), ranked[j].row.CreatedAt.ToTime()
		if !left.Equal(right) {
			return left.After(right)
		}
		return ranked[i].row.FileID > ranked[j].row.FileID
	})
	page, pageSize = normalizeCinemaPage(page, pageSize, 20, 100)
	start := (page - 1) * pageSize
	if start > len(ranked) {
		start = len(ranked)
	}
	end := start + pageSize
	if end > len(ranked) {
		end = len(ranked)
	}
	pageRows := make([]cinemaVideoRow, 0, end-start)
	for _, item := range ranked[start:end] {
		pageRows = append(pageRows, item.row)
	}
	videos, err := s.buildVideoItems(ctx, userID, tree, pageRows)
	if err != nil {
		return nil, err
	}
	return &response.CinemaRelatedResponse{
		Videos: videos, Total: len(ranked), Page: page, PageSize: pageSize, HasMore: end < len(ranked),
	}, nil
}

func (s *CinemaService) ValidateVideo(ctx context.Context, userID string, rootID int, fileID string) error {
	_, err := s.VideoDetail(ctx, userID, rootID, fileID)
	if err != nil {
		return fmt.Errorf("视频校验失败: %w", err)
	}
	return nil
}

func (s *CinemaService) RandomVideo(ctx context.Context, userID string, rootID int) (*response.CinemaVideoDetailResponse, error) {
	tree, err := s.loadTree(ctx, userID, rootID)
	if err != nil {
		return nil, err
	}

	// Get all video file IDs for random selection
	var fileIDs []string
	if err := s.playableVideoQuery(ctx, userID, tree.ids).
		Select("uf.uf_id").Scan(&fileIDs).Error; err != nil {
		return nil, err
	}

	if len(fileIDs) == 0 {
		return nil, errors.New("影视库中没有可播放视频")
	}

	// Select random file ID
	randomIndex := rand.Intn(len(fileIDs))
	randomFileID := fileIDs[randomIndex]

	// Get full video detail
	return s.VideoDetail(ctx, userID, rootID, randomFileID)
}
