// Package dto defines request and response payloads. Request structs carry
// `binding` tags consumed by Gin's validator; failures are rendered as the
// structured validation errors defined in the response package.
package dto

// PageQuery is the shared pagination/search query for list endpoints.
type PageQuery struct {
	Page    int    `form:"page" json:"page"`
	PerPage int    `form:"per_page" json:"per_page"`
	Search  string `form:"search" json:"search"`
}

// Normalize clamps pagination inputs into sane bounds.
func (q *PageQuery) Normalize() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 {
		q.PerPage = 20
	}
	if q.PerPage > 100 {
		q.PerPage = 100
	}
}

// Offset is the SQL OFFSET for the current page.
func (q PageQuery) Offset() int { return (q.Page - 1) * q.PerPage }
