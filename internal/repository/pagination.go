package repository

type Pagination struct {
	Page     int
	PageSize int
}

func (p Pagination) Limit() int {
	if p.PageSize <= 0 {
		return 20
	}
	return p.PageSize
}

func (p Pagination) Offset() int {
	if p.Page <= 0 {
		return 0
	}
	return (p.Page - 1) * p.Limit()
}
