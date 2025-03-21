package domain

import "time"

type Order struct {
	ID           int
	CustomerName string
	TotalAmount  float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Items        []OrderItem
}

type OrderItem struct {
	ID        int
	OrderID   int
	ProductID int
	Quantity  int
	Subtotal  float64
}

type CreateOrderItemService struct {
	ProductID int `validate:"required"`
	Quantity  int `validate:"required,min=1"`
}

type CreateOrderService struct {
	CustomerName string `validate:"required"`
	Items        []CreateOrderItemService `validate:"required,dive,required"`
}
