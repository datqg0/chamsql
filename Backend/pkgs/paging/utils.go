package paging

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// FromContext extracts Pageable from gin context query params
// Usage: pageable := paging.FromContext(c)
func FromContext(c *gin.Context) *Pageable {
	page := DefaultPage
	limit := DefaultLimit

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	pageable := NewPageableWithParams(page, limit)

	// Parse sort params: ?sort=name:asc,created_at:desc
	if sortStr := c.Query("sort"); sortStr != "" {
		pageable.Sort = ParseSortString(sortStr)
	}

	return pageable
}

// ParseSortString parses sort string like "name:asc,created_at:desc"
func ParseSortString(s string) []Order {
	if s == "" {
		return nil
	}

	var orders []Order
	parts := strings.Split(s, ",")
	for _, part := range parts {
		fieldDir := strings.Split(strings.TrimSpace(part), ":")
		if len(fieldDir) >= 1 && fieldDir[0] != "" {
			order := Order{Property: fieldDir[0], Direction: ASC}
			if len(fieldDir) >= 2 && strings.ToLower(fieldDir[1]) == "desc" {
				order.Direction = DESC
			}
			orders = append(orders, order)
		}
	}
	return orders
}
