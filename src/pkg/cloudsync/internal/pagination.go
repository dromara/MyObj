package internal

// PageCursorFetch 拉取一页数据，返回条目、下一页游标（空表示没有更多）、错误
type PageCursorFetch[T any] func(cursor string, pageSize int) (items []T, nextCursor string, err error)

// FetchPageByCursor 跳转到第 page 页（1-based）并返回结果
func FetchPageByCursor[T any, R any](
	page, pageSize int,
	fetch PageCursorFetch[T],
	convert func(T) R,
) ([]R, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}

	cursor := ""
	for i := 0; i < page; i++ {
		items, next, err := fetch(cursor, pageSize)
		if err != nil {
			return nil, 0, err
		}
		if i == page-1 {
			result := make([]R, 0, len(items))
			for _, item := range items {
				result = append(result, convert(item))
			}
			total := (page-1)*pageSize + len(result)
			if next != "" {
				total = page*pageSize + 1
			}
			return result, total, nil
		}
		if next == "" {
			return nil, 0, nil
		}
		cursor = next
	}
	return nil, 0, nil
}
