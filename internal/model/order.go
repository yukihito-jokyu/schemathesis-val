package model

type OrderStatus string

const (
	StatusCreated  OrderStatus = "created"
	StatusPaid     OrderStatus = "paid"
	StatusCanceled OrderStatus = "canceled"
)

type Order struct {
	ID       string      `json:"id"`
	UserID   int         `json:"userId"`
	ItemIDs  []string    `json:"itemIds"`
	Quantity int         `json:"quantity"`
	Status   OrderStatus `json:"status"`
}

type CreateOrderRequest struct {
	UserID   *int      `json:"userId"`
	ItemIDs  *[]string `json:"itemIds"`
	Quantity *int      `json:"quantity"`
}
