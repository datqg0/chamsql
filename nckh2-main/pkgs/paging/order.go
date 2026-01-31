package paging

// Direction represents sort direction
type Direction string

const (
	ASC  Direction = "ASC"
	DESC Direction = "DESC"
)

// Order represents a sort order
type Order struct {
	Property  string    `json:"property"`
	Direction Direction `json:"direction"`
}

// NewOrder creates a new Order
func NewOrder(property string, direction Direction) Order {
	return Order{
		Property:  property,
		Direction: direction,
	}
}
