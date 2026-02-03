package paging

const (
	DefaultLimit = 10
	DefaultPage  = 1
	MaximumLimit = 200
)

// Pageable represents pagination parameters
type Pageable struct {
	Sort   []Order           `json:"sort,omitempty"`
	Page   int               `json:"page"`
	Offset int               `json:"offset"`
	Limit  int               `json:"limit"`
	Total  int64             `json:"total,omitempty"`
	Params map[string]string `json:"params,omitempty"`
}

// NewPageable creates a new Pageable with default values
func NewPageable() *Pageable {
	return &Pageable{
		Page:   DefaultPage,
		Limit:  DefaultLimit,
		Offset: 0,
	}
}

// NewPageableWithParams creates a Pageable with custom page and limit
func NewPageableWithParams(page, limit int) *Pageable {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaximumLimit {
		limit = MaximumLimit
	}

	return &Pageable{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

// GetOffset calculates and returns the offset
func (p *Pageable) GetOffset() int {
	if p.Offset < 0 {
		offset := (p.Page - 1) * p.Limit
		if offset < 0 {
			return 0
		}
		return offset
	}
	return p.Offset
}

// SetPage sets the page and recalculates offset
func (p *Pageable) SetPage(page int) *Pageable {
	p.Page = page
	p.Offset = (page - 1) * p.Limit
	return p
}

// SetLimit sets the limit
func (p *Pageable) SetLimit(limit int) *Pageable {
	if limit > MaximumLimit {
		limit = MaximumLimit
	}
	p.Limit = limit
	p.Offset = (p.Page - 1) * limit
	return p
}

// SetTotal sets the total count
func (p *Pageable) SetTotal(total int64) *Pageable {
	p.Total = total
	return p
}

// AddSort adds a sort order
func (p *Pageable) AddSort(property string, direction Direction) *Pageable {
	p.Sort = append(p.Sort, Order{Property: property, Direction: direction})
	return p
}
