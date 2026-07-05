package model

type ItemCategory string

const (
	CategoryBook ItemCategory = "book"
	CategoryFood ItemCategory = "food"
	CategoryTool ItemCategory = "tool"
)

type Item struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Category ItemCategory `json:"category"`
	Price    float64      `json:"price"`
}

type CreateItemRequest struct {
	Name     *string       `json:"name"`
	Category *ItemCategory `json:"category"`
	Price    *float64      `json:"price"`
}
