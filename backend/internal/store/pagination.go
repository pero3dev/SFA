package store

type Pagination struct {
	Page  int32
	Limit int32
}

func (p Pagination) Offset() int32 {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit
}
