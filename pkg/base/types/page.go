package types

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// Page is the page response contract
type Page[T any] struct {
	Items      []T    `json:"items"`
	TotalItems uint64 `json:"totalItems"`
}

// PageRequest is the contract of request page
type PageRequest struct {
	Page  uint16
	Size  uint16
	Order []Sort
}

// NewPageRequest returns a new page request pointer
func NewPageRequest(page uint16, size uint16, order []Sort) *PageRequest {
	return &PageRequest{page, size, order}
}

// GetOrder returns string contains concated order list
func (p *PageRequest) GetOrder() string {
	orders := make([]string, 0, len(p.Order))

	for _, order := range p.Order {
		orders = append(orders, fmt.Sprintf("%s %s", order.Field, order.Direction))
	}

	return strings.Join(orders, ", ")
}

// GetMongoOrder returns string contains concated order list
func (p *PageRequest) GetMongoOrder() bson.D {
	var sort bson.D

	for _, order := range p.Order {
		direction := -1
		if order.Direction == ASC {
			direction = 1
		}

		sort = append(sort, bson.E{
			Key:   order.Field,
			Value: direction,
		})
	}

	return sort
}
