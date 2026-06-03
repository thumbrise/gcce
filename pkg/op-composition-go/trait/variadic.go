package trait

import "github.com/thumbrise/op-universal-schema-go/schema"

const (
	VariadicID      = BaseId + "/variadic"
	VariadicComment = "Marks a dependency as the variadic parameter."
)

func NewVariadic() schema.Term {
	return schema.Term{
		ID:      VariadicID,
		Comment: VariadicComment,
	}
}
