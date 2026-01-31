package paging

type Paging struct {
	Page     int   `json:"page" form:"page"`
	PageSize int   `json:"pageSize" form:"pageSize"`
	Total    int64 `json:"total"`
}

func (p *Paging) Process() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 || p.PageSize > 100 {
		p.PageSize = 20
	}
}

func (p *Paging) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func (p *Paging) TotalPages() int {
	if p.Total == 0 {
		return 0
	}
	return int((p.Total + int64(p.PageSize) - 1) / int64(p.PageSize))
}
