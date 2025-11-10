package models

import (
	"time"

	productsv1 "github.com/yaninyzwitty/go-fx-v1/gen/products/v1"
)

type ProductResponse struct {
	ID            uint64    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	Currency      string    `json:"currency"`
	StockQuantity uint32    `json:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Map from gRPC Product to REST-friendly ProductResponse
func FromProtoProduct(p *productsv1.Product) ProductResponse {
	return ProductResponse{
		ID:            p.GetId(),
		Name:          p.GetName(),
		Description:   p.GetDescription(),
		Price:         p.GetPrice(),
		Currency:      p.GetCurrency(),
		StockQuantity: p.GetStockQuantity(),
		CreatedAt:     p.GetCreatedAt().AsTime(),
		UpdatedAt:     p.GetUpdatedAt().AsTime(),
	}
}

// Generic API wrapper for a single object
type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// List wrapper for paginated responses
type ListResponse[T any] struct {
	Success       bool   `json:"success"`
	Message       string `json:"message,omitempty"`
	Data          []T    `json:"data"`
	NextPageToken uint32 `json:"next_page_token,omitempty"`
	Error         string `json:"error,omitempty"`
}

// Convenience helpers
func SuccessResponse[T any](data T, msg string) *APIResponse[T] {
	return &APIResponse[T]{
		Success: true,
		Message: msg,
		Data:    data,
	}
}

func ErrorResponse[T any](errMsg string) *APIResponse[T] {
	return &APIResponse[T]{
		Success: false,
		Error:   errMsg,
	}
}

func ListSuccessResponse[T any](data []T, nextPageToken uint32, msg string) *ListResponse[T] {
	return &ListResponse[T]{
		Success:       true,
		Message:       msg,
		Data:          data,
		NextPageToken: nextPageToken,
	}
}
