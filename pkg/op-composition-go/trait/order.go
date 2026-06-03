package trait

import "github.com/thumbrise/op-universal-schema-go/schema"

const (
	OrderID      = BaseId + "/order"
	OrderComment = "Order index of operation in resolved composition."
)

func NewOrder(value int) schema.Term {
	return schema.Term{
		ID:      OrderID,
		Comment: OrderComment,
		Value:   value,
	}
}
