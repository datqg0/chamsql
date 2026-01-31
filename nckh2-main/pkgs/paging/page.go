package paging

// Page represents a paginated result with generics
type Page[T any] struct {
	Total        int64 `json:"total"`
	Page         int   `json:"page"`
	Items        []T   `json:"items"`
	LoadMoreAble bool  `json:"loadMoreAble"`
	PreLoadAble  bool  `json:"preLoadAble,omitempty"`
}

// NewPage creates a new Page from pageable and items
func NewPage[T any](pageable *Pageable, items []T) *Page[T] {
	loadMoreAble := false
	if pageable.Total > 0 {
		loadMoreAble = int64(pageable.GetOffset()+pageable.Limit) < pageable.Total
	}

	return &Page[T]{
		Total:        pageable.Total,
		Page:         pageable.Page,
		Items:        items,
		LoadMoreAble: loadMoreAble,
	}
}

// NewPageWithTotal creates a new Page with total, page number and items
func NewPageWithTotal[T any](total int64, page int, items []T) *Page[T] {
	return &Page[T]{
		Total: total,
		Page:  page,
		Items: items,
	}
}

// Empty returns an empty page
func Empty[T any]() *Page[T] {
	return &Page[T]{
		Total: 0,
		Page:  1,
		Items: []T{},
	}
}

// Of creates a page from items with total and page number
func Of[T any](items []T, total int64, page int) *Page[T] {
	return &Page[T]{
		Total: total,
		Page:  page,
		Items: items,
	}
}

// IsEmpty returns true if the page has no items
func (p *Page[T]) IsEmpty() bool {
	return len(p.Items) == 0
}

// HasNext returns true if there are more pages
func (p *Page[T]) HasNext() bool {
	return p.LoadMoreAble
}
