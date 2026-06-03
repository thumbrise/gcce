package trait

import "github.com/thumbrise/op-universal-schema-go/schema"

const (
	ImplementsID      = BaseId + "/implements"
	ImplementsComment = "Contract fqn that entity implements"
)

func NewImplements(contractFQN string) schema.Term {
	return schema.Term{
		ID:      ImplementsID,
		Comment: ImplementsComment,
		Value:   contractFQN,
	}
}
