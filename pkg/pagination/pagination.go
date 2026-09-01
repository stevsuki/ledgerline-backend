// Package pagination: normalizes the page/per_page parameters.
package pagination

const (
	DefaultPage    = 1
	DefaultPerPage = 10
	MaxPerPage     = 100
)

type Params struct {
	Page    int `form:"page" binding:"omitempty,min=1"`
	PerPage int `form:"per_page" binding:"omitempty,min=1,max=100"`
}

// Normalize keeps values inside a safe range.
func (p Params) Normalize() Params {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.PerPage < 1 {
		p.PerPage = DefaultPerPage
	}
	if p.PerPage > MaxPerPage {
		p.PerPage = MaxPerPage
	}
	return p
}

func (p Params) Limit() int  { return p.Normalize().PerPage }
func (p Params) Offset() int { n := p.Normalize(); return (n.Page - 1) * n.PerPage }

// TotalPages: rounded up.
func TotalPages(totalItems, perPage int) int {
	if perPage <= 0 || totalItems <= 0 {
		return 0
	}
	return (totalItems + perPage - 1) / perPage
}
